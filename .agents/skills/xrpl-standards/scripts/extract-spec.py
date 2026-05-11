#!/usr/bin/env python3
"""
Compress a raw XLS README.md into implementation-relevant content.

Keeps:
  - Frontmatter block (<pre>...</pre> or ---...---)
  - Section headings for kept sections
  - All markdown tables (field definitions)
  - Bullet list items in kept sections (failure conditions, state changes, flags)

Drops:
  - All prose paragraphs
  - Code blocks (examples)
  - Background, motivation, rationale, introduction, terminology, overview
  - Security considerations
  - FAQ, acknowledgements, references, appendix, examples, compliance
"""

import sys
import re

KEEP_KEYWORDS = frozenset({
    "transaction", "ledger object", "ledger entry", "field", "flags",
    "invariant", "error", "rpc", "api", "validation", "serialization",
    "amendment", "account", "offer", "escrow", "payment", "trust",
    "nft", "token", "credential", "bridge", "oracle", "permission",
    "fee", "reserve", "object type", "entry type", "format", "encoding",
    "failure condition", "state change", "request", "response",
    "preauth", "deposit", "object:", "transaction:", "rpc:", "on-ledger",
})

# Drop wins over keep — checked first
DROP_KEYWORDS = frozenset({
    "background", "motivation", "rationale", "introduction",
    "terminology", "basic flow", "overview", "abstract",
    "security consideration", "security", "trust assumption", "data privacy",
    "faq", "frequently asked", "acknowledgement", "acknowledgment",
    "reference", "appendix", "prior art", "alternative",
    "example", "compliance", "history", "summary", "changelog",
})

HEADING_RE = re.compile(r'^(#{1,6})\s+(.*)')


def section_name(heading_line: str) -> str:
    """Extract normalized heading text: strip # markers and leading numbers."""
    text = re.sub(r'^#+\s+', '', heading_line).strip()
    # Strip leading section numbers like "1.", "2.3.", "A.1:"
    text = re.sub(r'^[A-Za-z]?\d+[\d.]*[.:)]\s*', '', text).strip()
    return text.lower()


def classify_heading(line: str) -> tuple[int, bool | None]:
    """
    Returns (level, decision):
      True  → always keep
      False → always drop
      None  → inherit from parent
    """
    m = HEADING_RE.match(line)
    if not m:
        return (0, None)

    level = len(m.group(1))
    name = section_name(line)

    # Drop always wins
    if any(kw in name for kw in DROP_KEYWORDS):
        return (level, False)

    if any(kw in name for kw in KEEP_KEYWORDS):
        return (level, True)

    # Default: keep top-level (# and ##), inherit for deeper
    if level <= 2:
        return (level, None)
    return (level, None)


class SectionState:
    """Tracks keep/drop decision. Explicit drop cascades to all children."""

    def __init__(self):
        # (level, effective_keep, force_drop)
        # force_drop=True means this AND all children are dropped, no exceptions
        self._stack: list[tuple[int, bool, bool]] = [(0, True, False)]

    def push(self, level: int, explicit: bool | None) -> bool:
        # Pop entries at same or deeper level
        while self._stack and self._stack[-1][0] >= level:
            self._stack.pop()

        _, parent_keep, parent_force_drop = self._stack[-1] if self._stack else (0, True, False)

        if parent_force_drop:
            # Explicitly-dropped parent cascades: no child can override
            effective = False
            force_drop = True
        elif explicit is False:
            effective = False
            force_drop = True   # cascade to children
        elif explicit is True:
            effective = True
            force_drop = False
        else:
            # Inherit from parent
            effective = parent_keep
            force_drop = False

        self._stack.append((level, effective, force_drop))
        return effective


def parse(lines: list[str]) -> list[str]:
    output: list[str] = []
    in_frontmatter = False
    frontmatter_done = False
    frontmatter_fence_count = 0

    state = SectionState()
    current_section_keep = True
    current_section_lines: list[str] = []
    pending_heading: str | None = None

    def flush_section():
        nonlocal current_section_lines, pending_heading

        if not current_section_lines and pending_heading is None:
            return

        if current_section_keep:
            if pending_heading is not None:
                output.append(pending_heading)
            emit_kept_section(current_section_lines)
        else:
            # Dropped section: emit tables only (field defs are always useful)
            tables = extract_tables(current_section_lines)
            if tables and pending_heading is not None:
                output.append(pending_heading)
                output.extend(tables)

        pending_heading = None
        current_section_lines.clear()

    def extract_tables(section_lines: list[str]) -> list[str]:
        """Extract all markdown table blocks from a list of lines."""
        result: list[str] = []
        in_table = False
        for line in section_lines:
            if line.startswith('|'):
                in_table = True
                result.append(line)
            elif in_table:
                if line.strip() == '':
                    result.append('')
                    in_table = False
                else:
                    in_table = False
        # Trim trailing blank
        while result and result[-1].strip() == '':
            result.pop()
        return result

    def emit_kept_section(section_lines: list[str]):
        """
        Emit content from a kept section:
        - Markdown tables (always)
        - Bullet/numbered list items (failure conditions, state changes, etc.)
        - Blank lines between items
        Skip: prose paragraphs, code blocks
        """
        in_table = False
        in_code_block = False
        prev_was_content = False

        for line in section_lines:
            stripped = line.strip()

            # Code blocks: skip entirely
            if stripped.startswith('```') or stripped.startswith('~~~'):
                in_code_block = not in_code_block
                continue
            if in_code_block:
                continue

            # Tables
            if line.startswith('|'):
                in_table = True
                output.append(line)
                prev_was_content = True
                continue
            elif in_table:
                if stripped == '':
                    in_table = False
                    output.append('')
                    prev_was_content = False
                else:
                    in_table = False
                continue

            # Blank lines: preserve at most one
            if stripped == '':
                if prev_was_content:
                    output.append('')
                    prev_was_content = False
                continue

            # Blockquotes: keep — specs use these for important warnings/constraints
            if stripped.startswith('>'):
                output.append(line)
                prev_was_content = True
                continue

            # List items (bullet or numbered)
            if (stripped.startswith('- ') or
                    stripped.startswith('* ') or
                    stripped.startswith('+ ') or
                    re.match(r'^\d+\.\s', stripped)):
                output.append(line)
                prev_was_content = True
                continue

            # Nested list items (indented)
            if re.match(r'^\s{2,}[-*+]', line) or re.match(r'^\s{2,}\d+\.', line):
                output.append(line)
                prev_was_content = True
                continue

            # Skip: prose paragraphs, HTML, references, etc.

    i = 0
    while i < len(lines):
        line = lines[i]
        raw = line.rstrip()

        # --- Frontmatter ---
        if not frontmatter_done:
            if raw == '---' and frontmatter_fence_count == 0:
                in_frontmatter = True
                frontmatter_fence_count += 1
                output.append(raw)
                i += 1
                continue
            elif raw == '---' and in_frontmatter:
                in_frontmatter = False
                frontmatter_done = True
                output.append(raw)
                i += 1
                continue
            # <pre>...</pre> style frontmatter — check </pre> BEFORE generic in_frontmatter
            elif raw.strip() == '<pre>' and frontmatter_fence_count == 0:
                in_frontmatter = True
                frontmatter_fence_count += 1
                output.append(raw)
                i += 1
                continue
            elif raw.strip() == '</pre>' and in_frontmatter:
                in_frontmatter = False
                frontmatter_done = True
                output.append(raw)
                i += 1
                continue
            elif in_frontmatter:
                output.append(raw)
                i += 1
                continue

        # --- Section headings ---
        if HEADING_RE.match(raw):
            flush_section()
            level, explicit = classify_heading(raw)
            current_section_keep = state.push(level, explicit)
            pending_heading = raw
            i += 1
            continue

        # --- Section body ---
        current_section_lines.append(raw)
        i += 1

    flush_section()

    # Deduplicate consecutive blank lines and strip edges
    result: list[str] = []
    prev_blank = False
    for line in output:
        if line.strip() == '':
            if not prev_blank:
                result.append('')
            prev_blank = True
        else:
            result.append(line)
            prev_blank = False

    while result and result[0].strip() == '':
        result.pop(0)
    while result and result[-1].strip() == '':
        result.pop()

    return result


def main():
    if len(sys.argv) > 1:
        with open(sys.argv[1], 'r', encoding='utf-8') as f:
            content = f.read()
    else:
        content = sys.stdin.read()

    lines = content.splitlines()
    result = parse(lines)
    print('\n'.join(result))


if __name__ == '__main__':
    main()
