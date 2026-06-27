#!/usr/bin/env bash

set -euo pipefail

usage() {
  echo "Usage: scripts/bump-version.sh <major|minor|patch>" >&2
  echo "  DRY_RUN=1 to preview without creating or pushing the tag" >&2
}

die() {
  echo "error: $*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

bump_type="${1:-}"

if [[ $# -ne 1 ]]; then
  usage
  exit 2
fi

case "${bump_type}" in
  major | minor | patch) ;;
  *)
    usage
    exit 2
    ;;
esac

require_command git

repo_root="$(git rev-parse --show-toplevel)"
cd "${repo_root}"

if [[ -n "$(git status --porcelain)" ]]; then
  die "working tree must be clean"
fi

# Determine the default branch from the remote (fall back to "main"), then make
# sure we are on it. Releases must be cut from the default branch only.
default_branch="$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's@^origin/@@')"
default_branch="${default_branch:-main}"

current_branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "${current_branch}" != "${default_branch}" ]]; then
  die "must be on default branch '${default_branch}' (current: ${current_branch})"
fi

# The release workflow is triggered by pushing a tag (see .github/workflows/release.yml),
# and the version is injected from the git tag via GoReleaser. There is no version file
# in the source tree, so the current version is derived from the latest tag.
current_tag="$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-version:refname | head -n 1)"

if [[ -z "${current_tag}" ]]; then
  current_version="0.0.0"
else
  current_version="${current_tag#v}"
fi

if [[ ! "${current_version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  die "could not parse semver from latest tag: ${current_tag}"
fi

IFS=. read -r major minor patch <<< "${current_version}"

case "${bump_type}" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
esac

new_version="${major}.${minor}.${patch}"
tag_name="v${new_version}"

if git rev-parse --verify --quiet "refs/tags/${tag_name}" >/dev/null; then
  die "local tag already exists: ${tag_name}"
fi

if [[ "${DRY_RUN:-}" == "1" ]]; then
  echo "${current_version} -> ${new_version}"
  echo "branch: ${current_branch}"
  echo "tag: ${tag_name}"
  exit 0
fi

if git ls-remote --exit-code --tags origin "${tag_name}" >/dev/null 2>&1; then
  die "remote tag already exists: ${tag_name}"
fi

# Make sure the local default branch is up to date with the remote so the tag
# points at the right commit and the release builds what is actually on origin.
git fetch origin "${default_branch}"
if [[ "$(git rev-parse HEAD)" != "$(git rev-parse "origin/${default_branch}")" ]]; then
  die "local ${default_branch} is not in sync with origin/${default_branch} (run 'git pull' first)"
fi

git tag -a "${tag_name}" -m "Release ${tag_name}"
git push origin "${tag_name}"

echo "Pushed ${tag_name}. The Release workflow will build and publish the release."
