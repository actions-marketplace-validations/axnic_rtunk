package ignore_test

import (
	"testing"

	"github.com/axnic/rtunk/pkg/ignore"
)

// FuzzPathMatches throws arbitrary patterns and paths at the gitignore-style matcher. The patterns
// come from a configuration file and the paths from the repository, so neither is trusted: the
// matcher must never panic or loop, and must answer the same thing twice. The seed corpus runs on
// every `go test`; `go test -fuzz=FuzzPathMatches ./pkg/ignore` explores beyond it.
func FuzzPathMatches(f *testing.F) {
	for _, seed := range [][2]string{
		{"*.go", "main.go"},
		{"**/vendor/**", "a/vendor/b/c.go"},
		{"/build", "build/out"},
		{"docs/", "docs/a.md"},
		{"[a-c]?.txt", "bb.txt"},
		{"a/**/b", "a/x/y/b"},
		{"[", "["},
		{"**/**/**", ""},
		{"\\", "\\"},
	} {
		f.Add(seed[0], seed[1])
	}
	f.Fuzz(func(t *testing.T, pattern, rel string) {
		first := ignore.PathMatches(pattern, rel)
		if again := ignore.PathMatches(pattern, rel); first != again {
			t.Fatalf("PathMatches(%q, %q) is not deterministic: %v then %v", pattern, rel, first, again)
		}
	})
}
