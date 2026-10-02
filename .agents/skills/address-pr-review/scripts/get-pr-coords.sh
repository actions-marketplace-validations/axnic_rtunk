#!/usr/bin/env bash
# Print "OWNER REPO PR_NUMBER PR_AUTHOR" for a pull request. Read-only.
# Usage: get-pr-coords.sh [PR_NUMBER|PR_URL|BRANCH]   (default: PR of the current branch)
set -euo pipefail

gh pr view "$@" --json number,url,author \
  --jq '[(.url | capture("github.com/(?<o>[^/]+)/(?<r>[^/]+)/pull/") | .o, .r), .number, .author.login] | join(" ")'
