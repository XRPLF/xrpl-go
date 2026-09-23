# Releasing xrpl-go

Prepare normal releases on `main`, update the matching release snapshot branch, then publish through GitHub Actions from that branch. Prepare backports and hotfixes for older lines on their release branches.

[Choose a release](#choose-a-release) · [Release snapshots](#release-snapshots) · [Publish a release](#publish-a-release) · [Before publishing](#before-publishing)

| Module | Module directory | Changelog | Git tag |
| --- | --- | --- | --- |
| Core | `.` | `CHANGELOG.md` | `vX.Y.Z` |
| Confidential | `confidential/` | `confidential/CHANGELOG.md` | `confidential/vX.Y.Z` |

Both changelogs use version headings without a directory prefix, such as `## [v0.1.0]`.

For v0 releases, the minor component identifies breaking release lines. Use the patch component for compatible features and fixes. Each module advances its version independently.

## Choose a release

| Change | Release order |
| --- | --- |
| Core-only feature or fix | Core only |
| Confidential builder or native-library update using existing core APIs | Confidential only |
| Confidential change that requires a new core API or protocol size | Core first, then confidential |

When confidential needs a newer core API, publish core first. Then, from `confidential/`, update the dependency:

```bash
GOWORK=off go get github.com/Peersyst/xrpl-go@vX.Y.Z
GOWORK=off go mod tidy
```

Replace `vX.Y.Z` with the released core version. Test against that published dependency and merge the module changes into `main` before preparing the confidential release snapshot.

Core has no dependency on confidential. A core release does not need a matching confidential release. Versions and tags can point to different commits.

## Release snapshots

`main` contains ongoing development and normal release preparation. Release branches are low-activity snapshots, not parallel development branches. For normal releases, prepare code, dependencies, and changelog entries on `main`, then update the snapshot.

**Backports and hotfixes for an older release line are the exception.** Prepare them from that line's branch, not from the current `main`, which may contain incompatible changes. Use a backport only after `main` has moved to a newer line. While `main` still targets a line, land the fix on `main` and make a normal release, so the snapshot branch can still fast-forward.

| Module line | Snapshot branch | Release tags |
| --- | --- | --- |
| Core 0.3 | `v0.3.x` | `v0.3.0`, `v0.3.1`, etc. |
| Confidential 0.1 | `confidential/v0.1.x` | `confidential/v0.1.0`, `confidential/v0.1.1`, etc. |

Before publishing, advance the matching branch to the prepared commit on `main`. Create the branch at that commit if this is its first release. Use a fast-forward update so the snapshot points to the same commit, without a branch-only merge commit.

Leave the branch at the published commit between releases. This gives the release Action a stable source while development continues on `main`. The branch records the latest release for its line, while each immutable tag identifies one specific release.

Keep older branches alive while their lines are supported. They preserve the released baseline for backports and hotfixes. Do not advance an older branch to incompatible changes from `main`.

Protect release branches against deletion and force pushes. Backports can legitimately make an older branch diverge from `main`. Do not reset it to restore an exact snapshot. Investigate unexpected divergence before publishing, and never move an existing release tag.

### Backports and hotfixes

For example, to fix confidential `0.1.x` after `main` has moved to a newer line:

1. Create a short-lived fix branch from `confidential/v0.1.x`.
2. Apply only the required fix and regression test. If the fix already exists on `main`, use `git cherry-pick -x` to record its origin and adapt it as needed.
3. Prepare the next patch version's notes in `confidential/CHANGELOG.md` on the fix branch. Keep the changes compatible with the older API and dependencies.
4. Run the affected tests and lint against that line. Use `GOWORK=off` to check published dependencies, and run relevant integration tests.
5. Open a PR into `confidential/v0.1.x`, not `main`. After review and merge, run the release Action from `confidential/v0.1.x` with the **confidential** module and the new patch version.
6. Confirm the release tag points to the intended commit. Apply the fix to `main` and other supported lines when affected. A fix specific to the older line does not need an equivalent change on `main`.

Use the same process for core branches such as `v0.3.x`. These releases do not include a full merge from `main`.

## Publish a release

The following procedure is for normal releases. For an older line, use the [backport and hotfix procedure](#backports-and-hotfixes) instead.

1. **Prepare everything on `main`.** Merge the intended code, dependency updates, and release notes. Put the selected module's notes under the expected `## [vX.Y.Z]` heading in its changelog. Keep `[Unreleased]` only for changes intended for a later release.
2. **Verify the release commit.** Complete the [pre-release checks](#before-publishing) and record the prepared `main` commit. Confirm that its changes are compatible with the intended release line.
3. **Update the snapshot.** Advance `v0.3.x` or `confidential/v0.1.x`, as appropriate, to that exact commit. Do not make further edits on the release branch. If preparation needs a correction, land it on `main` first and update the snapshot again.
4. **Open the Action.** Go to **Actions -> Release -> Run workflow** in GitHub. Select the snapshot branch, not `main`, choose **core** or **confidential**, and enter the version, such as `v0.3.1` or `v0.1.0-rc1`.
5. **Publish and confirm.** Run the workflow. It validates module metadata, tests the selected module, and creates the release. Confirm that the tag and snapshot branch point to the prepared commit. Leave the branch there until the next release for that line.

**Enter `v0.1.0`, not `confidential/v0.1.0`.** The workflow adds the prefix. Prerelease suffixes produce GitHub prereleases. Existing tags are rejected, not moved.

The workflow checks published dependencies with `GOWORK=off`. A local workspace cannot satisfy a missing core release.

Confidential releases use `--latest=false` because GitHub has one Latest release per repository. This does not prevent Go from finding the latest confidential version by its prefixed tags.

## Before publishing

- [ ] All intended changes and versioned release notes are committed on `main` for a normal release, or on the target release branch for a backport or hotfix.
- [ ] The version, module, and snapshot branch match the intended release line.
- [ ] Normal CI checks pass for the prepared commit.
- [ ] Module paths are unchanged, unless the release includes an explicit path migration.
- [ ] Examples and tests that import confidential remain inside its module.
- [ ] Published `go.mod` files have no local replacements. Keep these in the ignored workspace.
- [ ] The selected module's tests pass against published dependencies.
- [ ] Native dependency updates stay under `confidential/deps/`. If headers change protocol sizes, core and the confidential minimum dependency are updated together.

The release workflow checks metadata and tests again before creating the tag.

Go releases are identified by Git tags and module paths. No extra compiled SDK asset or separate repository is required. GitHub source archives and repository clones still contain both modules, including the native bundles.
