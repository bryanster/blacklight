# Releasing Blacklight

How to cut and publish a release. The heavy lifting is done by
[`.github/workflows/release.yml`](../.github/workflows/release.yml); this
document describes what a maintainer does to trigger it and how to verify the
result.

## Prerequisites

- Push access to `github.com/bryanster/blacklight`.
- A checkout of `main` at the commit you want to release. CI must be green on
  that commit — branch protection enforces this.
- `CHANGELOG.md` has a `[<version>]` section with the release notes. The
  workflow extracts that section into the GitHub Release body.

## Steps

### 1. Update the changelog

Move content from `[Unreleased]` into a new `[<version>]` section, and set the
date. For the final `v1.0.0` release (M7-009), the date and commit SHA are
filled in that ticket. For earlier releases or release candidates:

```markdown
## [1.0.0] — 2026-08-15
```

In the same change, point the deployment examples at the new tag. The
**Deploy to Azure** button deploys `deploy/azure-arm/azuredeploy.json` straight
from `main`, so its `imageTag` default is what every new one-click install
runs:

- `deploy/azure-arm/azuredeploy.json` — `imageTag` `defaultValue`
- `deploy/azure-arm/azuredeploy.parameters.json` — `imageTag`
- `deploy/azure-terraform/variables.tf` and `terraform.tfvars.example` — `image_tag`
- `README.md` — the status line, and `deploy/azure-arm/README.md` — the
  parameter table and upgrade examples

`TestDeployPinsMatchLatestRelease` (`internal/version`) fails CI until the
first five match the newest `## [x.y.z]` section, so a release cannot merge
with a stale button. The button names an image that does not exist until the
tag's release workflow finishes, so push the tag promptly after merging.

Commit that change to `main` and push. CI must pass.

### 2. Tag

Create an **annotated** tag on `main`. Annotated tags carry a message, a
tagger, and a date; lightweight tags carry none of those and a release without
them is not credible.

```sh
git checkout main
git pull
git tag -a v1.0.0 -m "Blacklight v1.0.0"
```

Or for a prerelease:

```sh
git tag -a v1.0.0-rc.1 -m "Blacklight v1.0.0-rc.1"
```

Push the tag:

```sh
git push origin v1.0.0
```

### 3. Watch the workflow

The push triggers [`.github/workflows/release.yml`](../.github/workflows/release.yml).
Monitor it at **Actions → Release**. It:

1. Verifies the tag is on `main` (so CI passed).
2. Builds multi-arch images for `linux/amd64` and `linux/arm64`.
3. Pushes to `ghcr.io/bryanster/blacklight:<tag>`.
4. If the tag is a stable release (no `-rc`, `-alpha`, `-beta`), also pushes
   `:latest`.
5. Creates a GitHub Release with the matching `CHANGELOG.md` section and pull
   instructions.

### 4. Verify

Once the workflow is green:

```sh
# Pull the image
docker pull ghcr.io/bryanster/blacklight:v1.0.0

# Check the version stamp matches the tag
docker run --rm ghcr.io/bryanster/blacklight:v1.0.0 blacklight --version
# Expected: v1.0.0 (commit <sha>, built <rfc3339>)

# Smoke test it
IMAGE=ghcr.io/bryanster/blacklight IMAGE_TAG=v1.0.0 SKIP_BUILD=1 deploy/smoke.sh
```

Check the [GitHub Release](https://github.com/bryanster/blacklight/releases):

- The changelog section is present.
- The Docker pull command is in the body.
- The greenfield note is present.
- For a prerelease, the release is marked **Pre-release**.
- For a stable release, the release is marked **Latest**.

Check the [GHCR package](https://github.com/bryanster/blacklight/pkgs/container/blacklight):

- The tag is present.
- The multi-arch manifest lists both `linux/amd64` and `linux/arm64`.

### 5. Record completion

For the final ship ticket (M7-009), record image digests in the completion
notes:

```sh
docker buildx imagetools inspect ghcr.io/bryanster/blacklight:v1.0.0 --raw | sha256sum
```

## Version policy

- **Product version:** `v1.0.0` (semver). Planning documents refer to the
  rebuild as "v2" but that is historical; tags, GHCR tags, and
  `internal/version` use `v1.x.y`.
- **Prereleases:** `v1.0.0-rc.1`, `v1.0.0-alpha.1`. They do not move `:latest`.
- **`latest`:** tracks the most recent stable (non-prerelease) release.

## Rollback

If a release is bad:

1. **Do not delete the tag or the GitHub Release.** That breaks anyone who
   already pulled the image — they get "manifest unknown."
2. Cut a new patch release with the fix (e.g. `v1.0.1`).
3. If the release must be withdrawn immediately, edit the GitHub Release body
   to add a warning at the top, and mark it as a prerelease so `:latest` does
   not resolve to it. This is a stopgap — cut a fixed release as soon as
   possible.
