#!/usr/bin/env bash
# WRITE OPERATION: reply to a review thread, then resolve it. Never run by default; call it only
# after the user approved the push, once per thread.
# Usage: resolve-thread.sh THREAD_NODE_ID "MESSAGE"      (THREAD_NODE_ID starts with PRRT_)
set -euo pipefail

[[ $# -eq 2 && -n $1 && -n $2 ]] || { echo "usage: $0 THREAD_NODE_ID MESSAGE" >&2; exit 2; }
[[ $1 == PRRT_* ]] || { echo "error: '$1' is not a review thread id (PRRT_...)" >&2; exit 2; }
thread=$1 message=$2

# shellcheck disable=SC2016 # GraphQL variables, not shell
gh api graphql -f threadId="$thread" -f body="$message" -f query='
  mutation($threadId:ID!,$body:String!){
    addPullRequestReviewThreadReply(input:{pullRequestReviewThreadId:$threadId,body:$body}){comment{url}}
  }' --jq '.data.addPullRequestReviewThreadReply.comment.url'

# shellcheck disable=SC2016
gh api graphql -f threadId="$thread" -f query='
  mutation($threadId:ID!){resolveReviewThread(input:{threadId:$threadId}){thread{isResolved}}}' \
  --jq '.data.resolveReviewThread.thread.isResolved'
