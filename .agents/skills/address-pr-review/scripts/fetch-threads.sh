#!/usr/bin/env bash
# Print the unresolved, non-outdated review threads of a PR as a compact JSON array:
#   [{node_id, db_id, author, path, line, body}]
# Threads whose first comment is by the PR author are skipped. Read-only.
# Usage: fetch-threads.sh OWNER REPO PR_NUMBER
set -euo pipefail

[[ $# -eq 3 ]] || { echo "usage: $0 OWNER REPO PR_NUMBER" >&2; exit 2; }
owner=$1 repo=$2 pr=$3

# shellcheck disable=SC2016 # GraphQL variables, not shell
query='query($owner:String!,$repo:String!,$pr:Int!,$endCursor:String){
  repository(owner:$owner,name:$repo){pullRequest(number:$pr){
    author{login}
    reviewThreads(first:100,after:$endCursor){
      pageInfo{endCursor hasNextPage}
      nodes{id isResolved isOutdated comments(first:1){nodes{databaseId body path line author{login}}}}
    }}}}'

# --paginate follows pageInfo through the $endCursor variable; --slurp gathers one document per page.
gh api graphql --paginate --slurp -f query="$query" -f owner="$owner" -f repo="$repo" -F pr="$pr" |
  jq -c '
    (.[0].data.repository.pullRequest.author.login // "") as $me
    | [ .[].data.repository.pullRequest.reviewThreads.nodes[]
        | select(.isResolved | not) | select(.isOutdated | not)
        | . as $t | $t.comments.nodes[0] | select(. != null)
        | select((.author.login // "") != $me)
        | {node_id: $t.id, db_id: .databaseId, author: (.author.login // "unknown"),
           path, line, body} ]'
