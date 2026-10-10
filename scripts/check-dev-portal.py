"""Build portal Go packages in a disposable container, never execute them."""

import argparse
import json
import os
import re
import shutil
import subprocess
import tempfile
import uuid
from pathlib import Path

MODULE = "github.com/Peersyst/xrpl-go"
# Use a current supported toolchain, independent of the portal's toolchain directives.
IMAGE = "golang:1.26-bookworm@sha256:a688600ca24f8a4d3ca77f95b0dd40704a9fc787c826660eb7ba0b641b8b175d"


def prepare(portal, destination, version):
    """Copy Go source only. Ignore upstream build scripts and module directives."""
    if not re.fullmatch(r"v\d+\.\d+\.\d+", version):
        raise ValueError("expected a stable core release tag, such as v0.3.1")
    source = portal / "_code-samples"
    if not source.is_dir() or source.is_symlink():
        raise ValueError("portal _code-samples directory is missing or symlinked")
    packages = set()
    total = 0
    for directory, dirs, files in os.walk(source, followlinks=False):
        directory = Path(directory)
        for name in dirs + files:
            if (directory / name).is_symlink():
                raise ValueError("symlinks are not supported in _code-samples")
        dirs[:] = [name for name in dirs if name not in {".git", "vendor", "node_modules"}]
        for name in sorted(files):
            if not name.endswith(".go") or name.endswith("_test.go"):
                continue
            path = directory / name
            relative = path.relative_to(source)
            # Keep package arguments and report text predictable. Fail rather than skip.
            if not re.fullmatch(r"[A-Za-z0-9_./-]+", relative.as_posix()):
                raise ValueError("unsupported Go source path")
            size = path.stat().st_size
            total += size
            if size > 1024 * 1024 or total > 20 * 1024 * 1024:
                raise ValueError("Go source size limit exceeded")
            target = destination / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, target)
            packages.add("./" + relative.parent.as_posix())
    if not packages:
        raise ValueError("no Go examples found")
    if len(packages) > 500:
        raise ValueError("Go package count limit exceeded")
    (destination / "go.mod").write_text(
        "module portal-check.invalid/examples\n\ngo 1.26.0\n\n"
        f"require {MODULE} {version}\n\n"
        # A requirement alone can be upgraded by Go's minimal version selection.
        f"replace {MODULE} => {MODULE} {version}\n"
    )
    return sorted(packages)


BUILD_SCRIPT = """set -eu
mkdir -p /scratch/src /scratch/home /scratch/tmp
cp -R /source/. /scratch/src/
cd /scratch/src
failed=0
for package in "$@"; do
    printf '\\n=== %s ===\\n' "$package"
    if timeout --kill-after=10s 180s go build -mod=mod -p=2 -o /scratch/example "$package"; then
        printf 'PASS %s\\n' "$package"
    else
        printf 'FAIL %s\\n' "$package"
        failed=$((failed + 1))
    fi
done
printf '\\nChecked %s packages, %s failed.\\n' "$#" "$failed"
exit "$((failed != 0))"
"""


SECTION = re.compile(r"=== (\./[A-Za-z0-9_./-]+) ===")
LOCATION = re.compile(r"([A-Za-z0-9_./-]+\.go):(\d+)(?::(\d+))?: (.*)")
# Progress lines from module resolution are noise, not errors.
NOISE = re.compile(r"go: (downloading|finding|found|extracting) ")


def summarize(log, total):
    """Collect failed packages and their compiler errors from the build log."""
    failed, current, errors = [], None, []
    for line in log.splitlines():
        if match := SECTION.fullmatch(line):
            current, errors = match.group(1), []
        elif current and line == f"FAIL {current}":
            failed.append({"package": current[2:], "errors": errors[:10]})
            current = None
        elif current and line.startswith("\t") and errors:
            # Go indents continuation lines, such as "have" and "want" details.
            errors[-1]["message"] += " " + line.strip()
        elif current and line and not line.startswith("# ") and not NOISE.match(line):
            match = LOCATION.fullmatch(line)
            if match:
                file, row, column, message = match.groups()
                errors.append({"file": file, "line": int(row), "column": int(column or 0),
                               "message": message})
            else:
                errors.append({"message": line})
    return {"checked": total, "failed": failed[:100]}


def docker_command(source, name, packages):
    return [
        "docker", "run", "--rm", "--name", name,
        "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges",
        "--user=65534:65534", "--cpus=2", "--memory=3g", "--memory-swap=3g",
        "--pids-limit=256", "--ulimit=nofile=1024:1024",
        "--ulimit=fsize=67108864:67108864",
        "--tmpfs=/scratch:rw,nosuid,nodev,size=2g,mode=1777",
        "--mount", f"type=bind,source={source},target=/source,readonly",
        "--entrypoint=/usr/bin/env", IMAGE, "-i",
        "PATH=/usr/local/go/bin:/usr/bin:/bin", "HOME=/scratch/home",
        "TMPDIR=/scratch/tmp", "GOCACHE=/scratch/cache", "GOPATH=/scratch/go",
        "CGO_ENABLED=0", "GOWORK=off", "GOENV=off", "GOTOOLCHAIN=local",
        "GOPROXY=https://proxy.golang.org", "GOSUMDB=sum.golang.org",
        "GOVCS=*:off", "GOTELEMETRY=off", "GOMAXPROCS=2",
        "/bin/bash", "-c", BUILD_SCRIPT, "portal-check", *packages,
    ]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("portal", type=Path)
    parser.add_argument("version")
    parser.add_argument("--report", type=Path, default=Path("portal-build.log"))
    parser.add_argument("--summary", type=Path, default=Path("portal-summary.json"))
    args = parser.parse_args()
    name = "portal-check-" + uuid.uuid4().hex
    # The parent writes the report. The container cannot access it or GitHub files.
    with args.report.open("w") as report, tempfile.TemporaryDirectory() as scratch:
        source = Path(scratch)
        try:
            packages = prepare(args.portal.resolve(), source, args.version)
            source.chmod(0o755)
            report.write(f"xrpl-go: {args.version}\nGo image: {IMAGE}\n")
            report.write(f"Discovered {len(packages)} Go package directories.\n")
            report.flush()
            result = subprocess.run(
                docker_command(source, name, packages),
                stdout=report, stderr=subprocess.STDOUT, timeout=2400, check=False,
            )
            report.flush()
            summary = summarize(args.report.read_text(errors="replace"), len(packages))
            args.summary.write_text(json.dumps(summary))
            return result.returncode
        except (OSError, ValueError, subprocess.TimeoutExpired) as error:
            report.write(f"Check failed: {error}\n")
            return 1
        finally:
            # Killing a Docker client does not stop its container.
            try:
                subprocess.run(
                    ["docker", "rm", "-f", name], stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL, timeout=30, check=False,
                )
            except (OSError, subprocess.TimeoutExpired):
                report.write("Container cleanup failed. Check the Docker daemon.\n")


if __name__ == "__main__":
    raise SystemExit(main())
