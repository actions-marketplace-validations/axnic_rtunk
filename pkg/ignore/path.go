package ignore

import (
	"path"
	"strings"
)

// PathMatches reports whether rel, a slash-separated path relative to the repository root, is
// covered by pattern, one gitignore-style glob of a trunk `lint.ignore` entry's paths:. The
// supported subset is:
//
//   - `*`, `?` and `[...]` match within one path segment, never across a "/";
//   - `**` as a whole segment matches any number of segments, zero included;
//   - a pattern with a "/" at its start or in its middle is anchored to the repository root, one
//     without (`*.log`, `.mise`) matches at any depth;
//   - a trailing "/" restricts the pattern to directories, so it covers what is under one;
//   - a pattern that matches a directory covers everything under it.
//
// Negation (a leading "!"), escaping and `.gitignore` files themselves are not supported: a "!"
// pattern simply never matches (config.Validate rejects it up front). Matching is by segments
// through path.Match, so a pattern it cannot parse (an unclosed "[") matches nothing.
func PathMatches(pattern, rel string) bool {
	dirOnly := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	if pattern == "" || strings.HasPrefix(pattern, "!") {
		return false
	}
	// Any "/" left (a leading one included) anchors the pattern; an unanchored one gets a leading **.
	anchored := strings.Contains(pattern, "/")
	segs := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	if !anchored {
		segs = append([]string{"**"}, segs...)
	}

	parts := strings.Split(rel, "/")
	for n := 1; n <= len(parts); n++ {
		if n == len(parts) && dirOnly {
			break // rel itself is a file: a directory-only pattern can only match what contains it
		}
		if matchSegments(segs, parts[:n]) {
			return true
		}
	}
	return false
}

// matchSegments reports whether pat matches parts in full, a "**" segment standing for any number
// of parts.
func matchSegments(pat, parts []string) bool {
	if len(pat) == 0 {
		return len(parts) == 0
	}
	if pat[0] == "**" {
		for i := 0; i <= len(parts); i++ {
			if matchSegments(pat[1:], parts[i:]) {
				return true
			}
		}
		return false
	}
	if len(parts) == 0 {
		return false
	}
	if ok, err := path.Match(pat[0], parts[0]); err != nil || !ok {
		return false
	}
	return matchSegments(pat[1:], parts[1:])
}
