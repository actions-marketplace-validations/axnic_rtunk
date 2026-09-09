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
// Resolve deliberately never fetches (that would be network access), a missing local source is a
// configuration error, not something Resolve silently tolerates into an empty result.
type SourceNotFoundError struct {
	SourceID string
	Path     string
}

func (e *SourceNotFoundError) Error() string {
	return fmt.Sprintf("config: plugin source %q: local path %s does not exist", e.SourceID, e.Path)
}

// UnsupportedSourceError reports that a plugins.sources entry is a git source (uri/ref). Resolve
// does not fetch git sources — that would be a `git clone`, i.e. network access, out of scope
// here (ROADMAP.md v0.2) — so it cannot resolve a config that depends on one.
type UnsupportedSourceError struct {
	SourceID string
	URI      string
	Ref      string
}

func (e *UnsupportedSourceError) Error() string {
	return fmt.Sprintf("config: plugin source %q: git sources are not fetched yet (uri=%s ref=%s)", e.SourceID, e.URI, e.Ref)
}

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
