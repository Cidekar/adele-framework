#!/usr/bin/env bash
#
# Generate a grouped release changelog from commit messages between the
# previous release tag and a target tag. Prints markdown to stdout.
#
# Usage: generate-changelog.sh <tag>
#   e.g. generate-changelog.sh v2.0.0
#
# The changelog is built purely from Conventional Commit subjects, grouped
# by type. Release-plumbing commits (version bumps, the release CI commits)
# are filtered out. Written for POSIX-ish bash and kept compatible with the
# macOS bash 3.2 on the self-hosted runner (no mapfile / associative arrays).
set -eu

REPO_URL="https://github.com/Cidekar/adele-framework"

CUR_TAG="${1:-}"
if [ -z "$CUR_TAG" ]; then
  echo "usage: $0 <tag>" >&2
  exit 1
fi

# Previous release tag: the newest vMAJOR.MINOR.PATCH tag that sorts strictly
# BELOW the current one (no -rc/prerelease suffix). Selecting the newest tag
# older than CUR_TAG keeps the range correct even when generating notes for a
# tag that is not the latest release.
PREV_TAG=$(git tag -l --sort=-v:refname \
  | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
  | awk -v cur="$CUR_TAG" '$0 == cur {seen=1; next} seen {print}' \
  | head -n1 || true)

if [ -z "$PREV_TAG" ]; then
  RANGE="$CUR_TAG"          # first release: whole history up to the tag
else
  RANGE="${PREV_TAG}..${CUR_TAG}"
fi

# Non-merge subjects in range, minus release-plumbing noise.
SUBJECTS=$(git log --no-merges --format='%s' "$RANGE" 2>/dev/null \
  | grep -vE '^chore: bump version|^chore: allow actions workflow|^write version tag v|^ci\(release\):' \
  || true)

# scrub_vendor: rewrite dependency-scanner references into neutral wording.
# We don't control the Dependabot name in the alerts, so present these as
# plain "security advisories"/"security fix" and drop the internal alert
# numbers. Order matters: handle the parenthetical and the "alerts #..."
# forms before the bare-word fallback.
scrub_vendor() {
  sed -E \
    -e 's/ ?\(Dependabot[^)]*\)/ (security fix)/g' \
    -e 's/Dependabot (Go )?alerts?[[:space:]]*#[0-9,#.-]*/security advisories/g' \
    -e 's/Dependabot/security advisories/g' \
    -e 's/  +/ /g' \
    -e 's/ +$//'
}

# emit_group <type-regex> <label>: print the section if it has any commits.
emit_group() {
  lines=$(printf '%s\n' "$SUBJECTS" \
    | grep -E "^$1(\(.*\))?!?:" 2>/dev/null \
    | sed -E 's/^[a-z]+(\(.*\))?!?: //' \
    | scrub_vendor \
    | sed 's/^/- /' || true)
  if [ -n "$lines" ]; then
    printf '%s\n%s\n\n' "$2" "$lines"
  fi
}

echo "## What's Changed"
echo ""
emit_group "feat"     "### 🚀 Features"
emit_group "fix"      "### 🐛 Fixes"
emit_group "build"    "### 🔧 Build & Dependencies"
emit_group "ci"       "### 👷 CI"
emit_group "test"     "### ✅ Tests"
emit_group "docs"     "### 📝 Docs"
emit_group "refactor" "### ♻️ Refactor"
emit_group "perf"     "### ⚡ Performance"
emit_group "chore"    "### 🧹 Chores"

if [ -n "$PREV_TAG" ]; then
  echo "**Full diff:** ${REPO_URL}/compare/${PREV_TAG}...${CUR_TAG}"
fi
