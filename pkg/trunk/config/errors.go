package config

import "fmt"

// ReadError reports that trunk.yaml or a plugin.yaml could not be read from disk — e.g. it
// doesn't exist, or permissions deny access. Unwrap returns the underlying error (typically an
// *fs.PathError), so callers can e.g. errors.Is(err, os.ErrNotExist).
type ReadError struct {
	Path string
	Err  error
}

func (e *ReadError) Error() string { return fmt.Sprintf("config: read %s: %v", e.Path, e.Err) }
func (e *ReadError) Unwrap() error { return e.Err }

// ParseError reports that trunk.yaml or a plugin.yaml was read successfully but is not valid
// YAML (or doesn't match the expected shape). Unwrap returns the underlying yaml decode error.
type ParseError struct {
	Path string
	Err  error
}

func (e *ParseError) Error() string { return fmt.Sprintf("config: parse %s: %v", e.Path, e.Err) }
func (e *ParseError) Unwrap() error { return e.Err }

// SourceNotFoundError reports that a plugins.sources entry's `local:` path does not exist on
// disk. A local source is expected to already be present — unlike a git source (uri/ref), which
// Resolve fetches itself — a missing local source is a configuration error, not something Resolve
// silently tolerates into an empty result.
type SourceNotFoundError struct {
	SourceID string
	Path     string
}

func (e *SourceNotFoundError) Error() string {
	return fmt.Sprintf("config: plugin source %q: local path %s does not exist", e.SourceID, e.Path)
}

// InvalidSourceError reports that a plugins.sources entry has neither `local:` nor `uri:` set, so
// Resolve has nothing to read or fetch for it (ARCHITECTURE.md: every source is one or the
// other).
type InvalidSourceError struct {
	SourceID string
}

func (e *InvalidSourceError) Error() string {
	return fmt.Sprintf("config: plugin source %q: neither local nor uri is set", e.SourceID)
}

// FetchError reports that a git plugin source's clone/checkout failed — no network, an
// unreachable uri, a ref that doesn't exist, and so on. Unwrap returns the underlying error
// (typically an *exec.ExitError with the git command's stderr).
type FetchError struct {
	SourceID string
	URI      string
	Ref      string
	Err      error
}

func (e *FetchError) Error() string {
	return fmt.Sprintf("config: plugin source %q: fetch %s@%s: %v", e.SourceID, e.URI, e.Ref, e.Err)
}
func (e *FetchError) Unwrap() error { return e.Err }

// DuplicateError reports that a resource was defined more than once while merging config: two
// plugin.yaml files (or two entries in the same one) declaring the same download/tool/lint/
// action/runtime name, or trunk.yaml's plugins.sources repeating an id. It is never fatal on its
// own — the later definition always fully replaces the earlier one, never a field-by-field merge
// — but it is always reported, since a repeated name is almost always a mistake.
type DuplicateError struct {
	Category string // "plugin source", "download", "tool", "lint", "action", or "runtime"
	Key      string // the id/name that was defined more than once
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("config: duplicate %s %q", e.Category, e.Key)
}

// ReferenceError reports that a definition names another resource, by id, that does not exist
// anywhere in the resolved config — e.g. a tool's `download:`/`runtime:` naming a download/
// runtime no plugin source defines, a linter's `tools:` naming an undefined tool, or trunk.yaml
// enabling a lint/action/runtime with no matching definition.
type ReferenceError struct {
	Category  string // category of the definition holding the reference: "tool", "runtime", "lint", "lint.enabled", ...
	Key       string // id/name of the definition holding the reference
	Field     string // the field holding the reference: "download", "runtime", "tools", "enabled"
	Reference string // the missing id/name
}

func (e *ReferenceError) Error() string {
	return fmt.Sprintf("config: %s %q: %s %q not found", e.Category, e.Key, e.Field, e.Reference)
}
