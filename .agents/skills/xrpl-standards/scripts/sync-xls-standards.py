#!/usr/bin/env python3
"""Sync local XRPL-Standards markdown references from upstream."""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Any
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


REPO = "XRPLF/XRPL-Standards"
ROOT_API = f"https://api.github.com/repos/{REPO}/contents"
RAW_BASE = f"https://raw.githubusercontent.com/{REPO}/master"


@dataclass
class SyncRecord:
    number: int
    remote_dir: str
    local_path: Path
    title: str = "Unknown"
    status: str = "Unknown"
    added: bool = False


def request_json(url: str, token: str | None = None) -> Any:
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "xrpl-go-skills-sync",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"

    req = Request(url, headers=headers)
    with urlopen(req, timeout=60) as response:
        text = response.read().decode("utf-8")

    try:
        return json.loads(text)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"Failed to parse JSON from {url}: {exc}") from exc


def request_text(url: str, token: str | None = None) -> str:
    headers = {
        "User-Agent": "xrpl-go-skills-sync",
    }
    if token:
        headers["Authorization"] = f"Bearer {token}"

    req = Request(url, headers=headers)
    with urlopen(req, timeout=60) as response:
        return response.read().decode("utf-8")


def extract_markdown(readme_text: str, extract_script: Path) -> str:
    proc = subprocess.run(
        [sys.executable, str(extract_script)],
        input=readme_text.encode("utf-8"),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.decode("utf-8").strip())

    return proc.stdout.decode("utf-8")


def parse_frontmatter(markdown: str) -> dict[str, str]:
    lines = markdown.splitlines()
    start = None
    end_token = None

    for index, line in enumerate(lines):
        token = line.strip()
        if token == "<pre>" or token == "---":
            start = index + 1
            end_token = "</pre>" if token == "<pre>" else "---"
            break

    if start is None:
        return {}

    end = None
    for index in range(start, len(lines)):
        if lines[index].strip() == end_token:
            end = index
            break

    if end is None:
        return {}

    result: dict[str, str] = {}
    for entry in lines[start:end]:
        if ":" not in entry:
            continue
        key, value = entry.split(":", 1)
        result[key.strip()] = value.strip().strip('"').strip("'")

    return result


def parse_xls_number_from_file(name: str) -> int | None:
    match = re.fullmatch(r"xls-(\d{4})\.md", name.lower())
    if not match:
        return None
    try:
        return int(match.group(1))
    except ValueError:
        return None


def read_local_references(reference_dir: Path) -> dict[int, Path]:
    known: dict[int, Path] = {}
    for file_path in reference_dir.glob("**/*.md"):
        if file_path.name == "INDEX.md":
            continue

        number = parse_xls_number_from_file(file_path.name)
        if number is None:
            continue

        # Keep the first file for a number when duplicates exist.
        known.setdefault(number, file_path)

    return known


def list_remote_directories(token: str | None) -> list[tuple[int, str]]:
    data = request_json(ROOT_API, token=token)
    pattern = re.compile(r"^XLS-(\d+)", re.IGNORECASE)

    items: list[tuple[int, str]] = []
    for item in data:
        if item.get("type") != "dir":
            continue

        name = item.get("name", "")
        match = pattern.match(name)
        if not match:
            continue

        try:
            items.append((int(match.group(1)), name))
        except ValueError:
            continue

    return sorted(items, key=lambda value: value[0])


def generate_index_entries(reference_dir: Path) -> dict[str, list[tuple[int, str, str, str]]]:
    by_category: dict[str, list[tuple[int, str, str, str]]] = {}

    for file_path in sorted(reference_dir.glob("**/*.md")):
        if file_path.name == "INDEX.md":
            continue

        number = parse_xls_number_from_file(file_path.name)
        if number is None:
            continue

        meta = parse_frontmatter(file_path.read_text(encoding="utf-8"))
        title = meta.get("title", "Unknown")
        status = meta.get("status", "Unknown")

        rel = file_path.relative_to(reference_dir)
        category = rel.parts[0]
        by_category.setdefault(category, []).append(
            (number, title, status, str(Path("references") / rel).replace("\\", "/"))
        )

    for rows in by_category.values():
        rows.sort(key=lambda value: value[0])

    return by_category


def render_index(entries: dict[str, list[tuple[int, str, str, str]]]) -> str:
    lines = ["# XRPL Standards Index", ""]

    category_order = [
        "identity",
        "tokens",
        "defi",
        "payments",
        "accounts",
        "data",
        "cross-chain",
        "smart-contracts",
        "core",
        "ecosystem",
        "UNCLASSIFIED",
    ]

    additional = sorted(set(entries.keys()) - set(category_order))
    ordered_categories = [name for name in category_order if name in entries]
    ordered_categories.extend(name for name in additional if name not in ordered_categories)

    for category in ordered_categories:
        rows = entries[category]
        if not rows:
            continue

        lines.append(f"## {category}")
        lines.append("| XLS | Title | Status | File |")
        lines.append("| ----- | ------- | -------- | ------ |")
        for number, raw_title, raw_status, file_path in rows:
            title = str(raw_title).replace("|", "\\|")
            lines.append(f"| {number} | {title} | {raw_status} | `{file_path}` |")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def safe_write(path: Path, content: str, dry_run: bool) -> bool:
    current = path.read_text(encoding="utf-8") if path.exists() else ""
    if current == content:
        return False

    if dry_run:
        return True

    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")
    return True


def sync(
    skill_dir: Path,
    token: str | None,
    dry_run: bool,
    *,
    use_full_text: bool = True,
) -> tuple[list[SyncRecord], list[SyncRecord], list[str]]:
    reference_dir = skill_dir / "references"
    extract_script = skill_dir / "scripts" / "extract-spec.py"
    index_path = reference_dir / "INDEX.md"

    local_map = read_local_references(reference_dir)
    updates: list[SyncRecord] = []
    added: list[SyncRecord] = []
    notes: list[str] = []

    remote_items = list_remote_directories(token)
    for number, remote_dir in remote_items:
        readme_url = f"{RAW_BASE}/{remote_dir}/README.md"

        try:
            readme_raw = request_text(readme_url, token=token)
        except HTTPError as exc:
            notes.append(f"XLS-{number}: failed readme fetch ({exc.code})")
            continue
        except URLError as exc:
            notes.append(f"XLS-{number}: failed readme fetch ({exc.reason})")
            continue

        if use_full_text:
            body = readme_raw
        else:
            try:
                body = extract_markdown(readme_raw, extract_script).rstrip("\n") + "\n"
            except Exception as exc:
                notes.append(f"XLS-{number}: extractor failed ({exc})")
                continue

        metadata = parse_frontmatter(readme_raw)
        title = metadata.get("title", "Unknown")
        status = metadata.get("status", "Unknown")

        target = local_map.get(number)
        added_xls = False
        if target is None:
            target = reference_dir / "UNCLASSIFIED" / f"xls-{number:04d}.md"
            added_xls = True

        changed = safe_write(target, body, dry_run)
        if changed:
            record = SyncRecord(
                number=number,
                remote_dir=remote_dir,
                local_path=target,
                title=title,
                status=status,
                added=added_xls,
            )
            if added_xls:
                added.append(record)
            else:
                updates.append(record)

            local_map[number] = target

    index_entries = generate_index_entries(reference_dir)
    index_content = render_index(index_entries)
    if safe_write(index_path, index_content, dry_run):
        notes.append("INDEX.md updated")

    return updates, added, notes


def main() -> int:
    parser = argparse.ArgumentParser(description="Update XRPL-Standards references from upstream")
    parser.add_argument(
        "--skill-dir",
        default=str(Path(__file__).resolve().parents[1]),
        help="Path to xrpl-standards skill directory",
    )
    parser.add_argument("--dry-run", action="store_true", help="Do not write files")
    parser.add_argument("--json", action="store_true", help="Print JSON summary")
    parser.add_argument(
        "--extract",
        action="store_true",
        help="Write extracted summary instead of full remote README.md content",
    )

    args = parser.parse_args()

    skill_dir = Path(args.skill_dir).resolve()
    if not skill_dir.is_dir():
        print(f"Error: skill directory not found: {skill_dir}", file=sys.stderr)
        return 1

    token = None
    for env_name in ("GITHUB_TOKEN", "GH_TOKEN", "TOKEN"):
        value = os.environ.get(env_name)
        if value:
            token = value
            break

    try:
        updates, added, notes = sync(skill_dir, token, args.dry_run, use_full_text=not args.extract)
    except Exception as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    summary = {
        "updated": [
            {
                "number": item.number,
                "title": item.title,
                "status": item.status,
                "remote_dir": item.remote_dir,
                "local_path": str(item.local_path),
            }
            for item in updates
        ],
        "added": [
            {
                "number": item.number,
                "title": item.title,
                "status": item.status,
                "remote_dir": item.remote_dir,
                "local_path": str(item.local_path),
            }
            for item in added
        ],
        "notes": notes,
    }

    if args.json:
        print(json.dumps(summary, indent=2))
    else:
        for line in notes:
            print(f"- {line}")
        print(f"Updated: {len(summary['updated'])}")
        print(f"Added: {len(summary['added'])}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
