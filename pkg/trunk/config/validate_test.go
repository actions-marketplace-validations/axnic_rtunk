package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axnic/rtunk/pkg/trunk/config"
)

// TestConfig_Validate_Empty: a zero-value Config (nil maps/slices throughout) has nothing to
// check and must not panic ranging over them.
func TestConfig_Validate_Empty(t *testing.T) {
	var cfg config.Config
	assert.NoError(t, cfg.Validate())
}

// TestConfig_Validate_EnabledLists covers every enabled-list category (runtimes/lint/actions),
// both the bare-id and `id@version`-pinned forms, present and missing.
func TestConfig_Validate_EnabledLists(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		wantCat     string
		wantMissing string // "" means Validate must return nil
	}{
		{
			name: "runtime enabled, missing definition",
			cfg: config.Config{Runtimes: config.CategoryConfig[config.Runtime]{
				Enabled: []string{"node"}, Definitions: map[string]config.Runtime{},
			}},
			wantCat: "runtime", wantMissing: "node",
		},
		{
			name: "runtime enabled, pinned version, missing definition",
			cfg: config.Config{Runtimes: config.CategoryConfig[config.Runtime]{
				Enabled: []string{"node@22.0.0"}, Definitions: map[string]config.Runtime{},
			}},
			wantCat: "runtime", wantMissing: "node",
		},
		{
			name: "runtime enabled, definition present",
			cfg: config.Config{Runtimes: config.CategoryConfig[config.Runtime]{
				Enabled: []string{"node@22.0.0"}, Definitions: map[string]config.Runtime{"node": {Type: "node"}},
			}},
		},
		{
			name: "lint enabled, missing definition",
			cfg: config.Config{Lint: config.LintConfig{CategoryConfig: config.CategoryConfig[config.Linter]{
				Enabled: []string{"eslint"}, Definitions: map[string]config.Linter{},
			}}},
			wantCat: "lint", wantMissing: "eslint",
		},
		{
			name: "lint enabled, definition present",
			cfg: config.Config{Lint: config.LintConfig{CategoryConfig: config.CategoryConfig[config.Linter]{
				Enabled: []string{"eslint"}, Definitions: map[string]config.Linter{"eslint": {Name: "eslint"}},
			}}},
		},
		{
			name: "action enabled, missing definition",
			cfg: config.Config{Actions: config.CategoryConfig[config.Action]{
				Enabled: []string{"commitlint"}, Definitions: map[string]config.Action{},
			}},
			wantCat: "action", wantMissing: "commitlint",
		},
		{
			name: "action enabled, definition present",
			cfg: config.Config{Actions: config.CategoryConfig[config.Action]{
				Enabled: []string{"commitlint"}, Definitions: map[string]config.Action{"commitlint": {ID: "commitlint"}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantMissing == "" {
				assert.NoError(t, err)
				return
			}
			var refErr *config.ReferenceError
			require.ErrorAs(t, err, &refErr)
			assert.Equal(t, tt.wantCat, refErr.Category)
			assert.Equal(t, "enabled", refErr.Field)
			assert.Equal(t, tt.wantMissing, refErr.Reference)
		})
	}
}

// TestConfig_Validate_References covers every by-id reference validateReferences checks (a
// tool's runtime/download, a runtime's download, a linter's tools), present and missing.
func TestConfig_Validate_References(t *testing.T) {
	tests := []struct {
		name      string
		cfg       config.Config
		wantCat   string
		wantKey   string
		wantField string
		wantRef   string // "" means Validate must return nil
	}{
		{
			name: "tool runtime, missing",
			cfg: config.Config{
				Tools:    map[string]config.Tool{"eslint": {Name: "eslint", Runtime: "node"}},
				Runtimes: config.CategoryConfig[config.Runtime]{Definitions: map[string]config.Runtime{}},
			},
			wantCat: "tool", wantKey: "eslint", wantField: "runtime", wantRef: "node",
		},
		{
			name: "tool runtime, present",
			cfg: config.Config{
				Tools:    map[string]config.Tool{"eslint": {Name: "eslint", Runtime: "node"}},
				Runtimes: config.CategoryConfig[config.Runtime]{Definitions: map[string]config.Runtime{"node": {Type: "node"}}},
			},
		},
		{
			name: "tool download, missing",
			cfg: config.Config{
				Tools:     map[string]config.Tool{"shellcheck": {Name: "shellcheck", Download: "shellcheck"}},
				Downloads: map[string]config.Download{},
			},
			wantCat: "tool", wantKey: "shellcheck", wantField: "download", wantRef: "shellcheck",
		},
		{
			name: "tool download, present",
			cfg: config.Config{
				Tools:     map[string]config.Tool{"shellcheck": {Name: "shellcheck", Download: "shellcheck"}},
				Downloads: map[string]config.Download{"shellcheck": {Name: "shellcheck"}},
			},
		},
		{
			name: "tool with neither runtime nor download set",
			cfg: config.Config{
				Tools: map[string]config.Tool{"standalone": {Name: "standalone"}},
			},
		},
		{
			name: "runtime download, missing",
			cfg: config.Config{
				Runtimes:  config.CategoryConfig[config.Runtime]{Definitions: map[string]config.Runtime{"node": {Type: "node", Download: "node"}}},
				Downloads: map[string]config.Download{},
			},
			wantCat: "runtime", wantKey: "node", wantField: "download", wantRef: "node",
		},
		{
			name: "runtime download, present",
			cfg: config.Config{
				Runtimes:  config.CategoryConfig[config.Runtime]{Definitions: map[string]config.Runtime{"node": {Type: "node", Download: "node"}}},
				Downloads: map[string]config.Download{"node": {Name: "node"}},
			},
		},
		{
			name: "lint tools, missing",
			cfg: config.Config{
				Lint:  config.LintConfig{CategoryConfig: config.CategoryConfig[config.Linter]{Definitions: map[string]config.Linter{"eslint": {Name: "eslint", Tools: []string{"eslint-bin"}}}}},
				Tools: map[string]config.Tool{},
			},
			wantCat: "lint", wantKey: "eslint", wantField: "tools", wantRef: "eslint-bin",
		},
		{
			name: "lint tools, present",
			cfg: config.Config{
				Lint:  config.LintConfig{CategoryConfig: config.CategoryConfig[config.Linter]{Definitions: map[string]config.Linter{"eslint": {Name: "eslint", Tools: []string{"eslint-bin"}}}}},
				Tools: map[string]config.Tool{"eslint-bin": {Name: "eslint-bin"}},
			},
		},
		{
			name: "lint files, missing",
			cfg: config.Config{
				Lint: config.LintConfig{
					CategoryConfig: config.CategoryConfig[config.Linter]{Definitions: map[string]config.Linter{"eslint": {Name: "eslint", Files: []string{"javascript"}}}},
					Files:          map[string]config.FileType{},
				},
			},
			wantCat: "lint", wantKey: "eslint", wantField: "files", wantRef: "javascript",
		},
		{
			name: "lint files, present",
			cfg: config.Config{
				Lint: config.LintConfig{
					CategoryConfig: config.CategoryConfig[config.Linter]{Definitions: map[string]config.Linter{"eslint": {Name: "eslint", Files: []string{"javascript"}}}},
					Files:          map[string]config.FileType{"javascript": {Name: "javascript"}},
				},
			},
		},
		{
			name: "file inherit, missing",
			cfg: config.Config{
				Lint: config.LintConfig{Files: map[string]config.FileType{"bazel": {Name: "bazel", Inherit: []string{"bazel-build"}}}},
			},
			wantCat: "file", wantKey: "bazel", wantField: "inherit", wantRef: "bazel-build",
		},
		{
			name: "file inherit, present",
			cfg: config.Config{
				Lint: config.LintConfig{Files: map[string]config.FileType{
					"bazel":       {Name: "bazel", Inherit: []string{"bazel-build"}},
					"bazel-build": {Name: "bazel-build"},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantRef == "" {
				assert.NoError(t, err)
				return
			}
			var refErr *config.ReferenceError
			require.ErrorAs(t, err, &refErr)
			assert.Equal(t, tt.wantCat, refErr.Category)
			assert.Equal(t, tt.wantKey, refErr.Key)
			assert.Equal(t, tt.wantField, refErr.Field)
			assert.Equal(t, tt.wantRef, refErr.Reference)
		})
	}
}

// TestConfig_Validate_JoinsMultipleErrors: every problem Validate finds is reported together, not
// just the first — callers get errors.Join's []error, one per missing reference.
func TestConfig_Validate_JoinsMultipleErrors(t *testing.T) {
	cfg := config.Config{
		Runtimes: config.CategoryConfig[config.Runtime]{Enabled: []string{"node"}, Definitions: map[string]config.Runtime{}},
		Lint:     config.LintConfig{CategoryConfig: config.CategoryConfig[config.Linter]{Enabled: []string{"eslint"}, Definitions: map[string]config.Linter{}}},
	}

	err := cfg.Validate()
	require.Error(t, err)

	joined, ok := err.(interface{ Unwrap() []error })
	require.True(t, ok, "errors.Join must return something implementing Unwrap() []error")
	assert.Len(t, joined.Unwrap(), 2)
}
