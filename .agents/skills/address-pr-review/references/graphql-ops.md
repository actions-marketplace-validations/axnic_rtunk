# GraphQL operations for PR review

All calls go through `gh api graphql`; the scripts in `../scripts/` wrap them.

## Coordinates

`scripts/get-pr-coords.sh [PR]` runs `gh pr view [PR] --json number,url,author` and prints
`OWNER REPO PR_NUMBER PR_AUTHOR` on one line (owner and repo come from the PR URL, so it is
correct for PRs opened from forks).

## Fetch review threads (read)

```graphql
query ($owner: String!, $repo: String!, $pr: Int!, $endCursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      author {
        login
      }
      reviewThreads(first: 100, after: $endCursor) {
        pageInfo {
          endCursor
          hasNextPage
        }
        nodes {
          id
          isResolved
          isOutdated
          comments(first: 1) {
            nodes {
              databaseId
              body
              path
              line
              author {
                login
              }
            }
          }
        }
      }
    }
  }
}
```

`gh api graphql --paginate --slurp` follows `pageInfo` as long as the cursor variable is named
exactly `$endCursor`; `--slurp` returns one document per page. `fetch-threads.sh` then keeps
threads with `isResolved:false`, `isOutdated:false` and a first comment not written by the PR
author, and emits:

```json
[{ "node_id": "PRRT_...", "db_id": 123456, "author": "reviewer", "path": "pkg/x/y.go", "line": 42, "body": "..." }]
```

`node_id` is the thread id (`PRRT_...`), `db_id` the first comment's numeric id.

## Reply and resolve (write)

`scripts/resolve-thread.sh THREAD_NODE_ID MESSAGE` runs two mutations, in order:

```graphql
mutation ($threadId: ID!, $body: String!) {
  addPullRequestReviewThreadReply(input: { pullRequestReviewThreadId: $threadId, body: $body }) {
    comment {
      url
    }
  }
}
mutation ($threadId: ID!) {
  resolveReviewThread(input: { threadId: $threadId }) {
    thread {
      isResolved
    }
  }
}
```

The reply uses the GraphQL thread id, so no REST comment id or PR number is needed. The script
refuses ids that do not start with `PRRT_` and exits 2 on bad arguments.
