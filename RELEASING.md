# Releasing xrpl-go

The repository contains two independently versioned Go modules. Publishing one does not require publishing the other.

| Module | Module directory | Changelog | Git tag |
| --- | --- | --- | --- |
| Core | `.` | `CHANGELOG.md` | `vX.Y.Z` |
| Confidential | `confidential/` | `confidential/CHANGELOG.md` | `confidential/vX.Y.Z` |

Both changelogs use version headings without a directory prefix, such as `## [v0.1.0]`.

For v0 releases, the minor component identifies breaking release lines. Use the patch component for compatible features and fixes. Each module advances its version independently.

## Releases

| Change | Release order |
| --- | --- |
| Core-only feature or fix | Core only |
| Confidential builder or native-library update using existing core APIs | Confidential only |
| Confidential change that requires a new core API or protocol size | Core first, then confidential |

When confidential needs a newer core API, update its minimum requirement with `GOWORK=off go get github.com/Peersyst/xrpl-go@<published-version>`, then run `GOWORK=off go mod tidy` and its tests. Commit the changes before release.

Core has no dependency on confidential. A core release does not need a matching confidential version or an empty confidential release. Versions and tags can point to different commits.

## Use the release picker

1. Ensure the normal CI checks pass for the intended release commit.
2. Put the selected module's release notes under `## [vX.Y.Z]` in its own changelog. Keep `[Unreleased]` for later work.
3. Open **Actions -> Release -> Run workflow**.
4. Select the release branch or ref, choose **core** or **confidential**, and enter the version, such as `v0.3.1` or `v0.1.0-rc1`.
5. Run the workflow. It checks out the selected commit, validates module metadata, tests the selected module, and creates the release.

Enter `v0.1.0`, not `confidential/v0.1.0`. The workflow adds the prefix. Prerelease suffixes produce GitHub prereleases. Existing tags are rejected and are not moved.

Confidential releases use `--latest=false` because GitHub has one Latest release per repository. This does not prevent Go from finding the latest confidential version by its prefixed tags.

## Before publishing

- Keep module paths unchanged unless the release explicitly includes a path migration.
- Keep examples and tests that import confidential inside its module.
- Keep local replacements in the ignored workspace, not in published `go.mod` files.
- Run the selected module's tests. The release workflow also checks its metadata and tests against published dependencies before creating a tag.
- Keep native dependency updates under `confidential/deps/`. If headers change protocol sizes, update core and the confidential minimum dependency together.

Go releases are identified by Git tags and module paths. No extra compiled SDK asset or separate repository is required. GitHub source archives and repository clones still contain both modules, including the native bundles.
