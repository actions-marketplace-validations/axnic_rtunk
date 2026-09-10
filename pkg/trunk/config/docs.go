// Package config parses and resolves trunk.yaml — the two-layer model described in
// ARCHITECTURE.md.
//
// Resolve reads trunk.yaml, and every plugin.yaml under its local plugin sources' category dirs,
// each into its own raw shape (as YAML naturally decodes it: lists, not maps), then folds every
// list into the Config's maps, keyed by each resource's own id/name. A repeated key is never
// field-by-field merged — the later definition fully replaces the earlier one — but every such
// collision is recorded as a *DuplicateError.
//
// Config.Validate then checks that everything referenced by id (a tool's download/runtime, a
// linter's tools, trunk.yaml's own enabled lists) actually exists among what was merged,
// recording a *ReferenceError for anything dangling. It is a separate step, not run by Resolve
// itself, so callers that only need what was actually read can skip it.
//
// Every recorded error (duplicates and dangling references) is returned/joined via errors.Join,
// so a single Resolve or Validate call reports every problem instead of one per run.
//
// Git plugin sources (uri/ref) are fetched via the system git binary (init/fetch --depth 1/
// checkout, so both tags and SHAs work) into a throwaway temp dir, parsed, then discarded — what
// persists on disk is a cache of the parsed definitions, keyed by uri+ref, since a pinned ref
// never changes content. A cache file that fails to decode (corrupted, or from an incompatible
// rtunk version) is dropped and regenerated from the network rather than treated as fatal. A
// source with neither local nor uri set is a config error, reported as *InvalidSourceError.
package config
