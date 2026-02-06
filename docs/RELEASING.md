# Tagging and releasing

This document explains how to publish a release of **cowardly** (create a version tag and publish binaries via GitHub Actions).

## Overview

- **Tagging** is done manually (or via the GitHub UI). There is no workflow that creates tags for you.
- **Releasing** is automated: when you push a tag matching `v*`, the [release workflow](../.github/workflows/release.yml) runs, builds the binary for each platform, packs each into a `.tar.gz` with CHANGELOG.md, LICENSE, and README.md, and creates a [GitHub Release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) with those archives attached.

## Tag format

Use **semantic version** tags so the workflow picks them up:

- `v1.0.0` — major release
- `v0.2.1` — minor/patch (e.g. pre-1.0)
- `v2.0.0-beta.1` — optional pre-release suffix

The workflow triggers on any tag that starts with `v` (pattern `v*`).

## How to release

### 1. Prepare the release

- Ensure `main` (or your default branch) has the changes you want in this release.
- Run tests and lint locally: `make test`, `make lint`, `make format-check`, `make lint-yaml`.
- Optionally update a changelog or version note in the repo.

### 2. Create and push the tag

From your repo root, on the branch you want to release (e.g. `main`):

**Using `make bump-version` (recommended):**

```bash
make bump-version           # bump patch (e.g. v0.2.3 → v0.2.4)
make bump-version PART=minor  # bump minor (e.g. v0.2.3 → v0.3.0)
make bump-version PART=major  # bump major (e.g. v0.2.3 → v1.0.0)
```

This reads the latest tag, bumps the version, creates an annotated tag, and pushes it. The release workflow triggers on push.

**Manual tagging:**

```bash
# Create an annotated tag (recommended; stores date and message)
git tag -a v1.0.0 -m "Release v1.0.0"

# Push the tag to the remote
git push origin v1.0.0
```

Lightweight tag (no message):

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 3. Let the workflow run

- Go to the **Actions** tab of the repository and open the **release** workflow run for your tag.
- The job will:
  1. Check out the code
  2. Set up Go (version from `go.mod`)
  3. For **darwin/amd64** (Intel) and **darwin/arm64** (Apple Silicon): build `cowardly`, then create a `.tar.gz` containing the executable, CHANGELOG.md, LICENSE, and README.md
  4. Create a GitHub Release for the tag and attach the archives
  5. Generate release notes from the tag

### 4. Download or share the release

- Open the **Releases** page of the repository.
- The new release will appear with the tag name (e.g. `v1.0.0`) and two assets:
  - `cowardly_v1.0.0_darwin_x86_64.tar.gz` (Intel Mac)
  - `cowardly_v1.0.0_darwin_arm64.tar.gz` (Apple Silicon)

Asset names follow the format **`cowardly_v{VERSION}_{OS}_{ARCH}.tar.gz`** (e.g. for future Linux: `cowardly_v1.0.0_linux_x86_64.tar.gz`). Each archive contains a **single top-level directory** (e.g. `cowardly_v1.0.0_darwin_arm64/`) with the `cowardly` executable, CHANGELOG.md, LICENSE, and README.md inside it.

Users can download and run, e.g. for Apple Silicon:

```bash
tar xzf cowardly_v1.0.0_darwin_arm64.tar.gz
cd cowardly_v1.0.0_darwin_arm64
chmod +x cowardly
./cowardly
```

(Or move the `cowardly` binary to a directory in `PATH`.)

## Releasing from the GitHub UI

You can also create a release (and tag) from the GitHub website:

1. Go to **Releases** → **Draft a new release**.
2. Click **Choose a tag**, type a new tag (e.g. `v1.0.0`), and select **Create new tag**.
3. Add a title and description if you like.
4. Publish the release.

**Note:** Creating only a tag in the UI (without publishing a release) will still trigger the release workflow, which will then create the release and attach the built archives. So either “publish release” with a new tag or “create tag and push” from the CLI will result in the same automated build and release.

## What the workflow does not do

- It does **not** create tags. You (or the GitHub UI) create the tag; the workflow runs when the tag is pushed.
- It does **not** run on every push to `main`. Only tag pushes matching `v*` trigger it.
- It only builds for **macOS** (darwin). Other OS/arch (e.g. Linux, Windows) are not built yet; when added, archives will follow the same naming: `cowardly_v{VERSION}_{OS}_{ARCH}.tar.gz` or `.zip` for Windows.

## Troubleshooting

- **Workflow didn’t run** — Ensure the tag was pushed to the same repo (`git push origin v1.0.0`). Check the **Actions** tab for the run.
- **Release exists but no assets** — Check the workflow logs for the "Build and pack" and “Create Release” steps; the job needs `contents: write` (already set in the workflow).
- **Wrong Go version** — The workflow uses `go-version-file: go.mod`; the Go version in `go.mod` is used to run the build.

## See also

- [.github/workflows/release.yml](../.github/workflows/release.yml) — Release workflow definition
- [.github/workflows/README.md](../.github/workflows/README.md) — Overview of all workflows
