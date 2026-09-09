package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_Validate_Empty: a zero-value Config (nil maps/slices throughout) has nothing to
// check and must not panic ranging over them.
func TestConfig_Validate_Empty(t *testing.T) {
	var cfg Config
	assert.NoError(t, cfg.Validate())
}

// TestConfig_Validate_EnabledLists covers every enabled-list category (runtimes/lint/actions),
// both the bare-id and `id@version`-pinned forms, present and missing.
func TestConfig_Validate_EnabledLists(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantCat     string
		wantMissing string // "" means Validate must return nil
	}{
		{
			name: "runtime enabled, missing definition",
			cfg: Config{Runtimes: CategoryConfig[Runtime]{
				Enabled: []string{"node"}, Definitions: map[string]Runtime{},
			}},
			wantCat: "runtime", wantMissing: "node",
		},
		{
			name: "runtime enabled, pinned version, missing definition",
			cfg: Config{Runtimes: CategoryConfig[Runtime]{
				Enabled: []string{"node@22.0.0"}, Definitions: map[string]Runtime{},
			}},
			wantCat: "runtime", wantMissing: "node",
		},
		{
			name: "runtime enabled, definition present",
			cfg: Config{Runtimes: CategoryConfig[Runtime]{
				Enabled: []string{"node@22.0.0"}, Definitions: map[string]Runtime{"node": {Type: "node"}},
			}},
		},
		{
			name: "lint enabled, missing definition",
			cfg: Config{Lint: CategoryConfig[Linter]{
				Enabled: []string{"eslint"}, Definitions: map[string]Linter{},
			}},
			wantCat: "lint", wantMissing: "eslint",
		},
		{
			name: "lint enabled, definition present",
			cfg: Config{Lint: CategoryConfig[Linter]{
				Enabled: []string{"eslint"}, Definitions: map[string]Linter{"eslint": {Name: "eslint"}},
			}},
		},
		{
			name: "action enabled, missing definition",
			cfg: Config{Actions: CategoryConfig[Action]{
				Enabled: []string{"commitlint"}, Definitions: map[string]Action{},
			}},
			wantCat: "action", wantMissing: "commitlint",
		},
		{
			name: "action enabled, definition present",
			cfg: Config{Actions: CategoryConfig[Action]{
				Enabled: []string{"commitlint"}, Definitions: map[string]Action{"commitlint": {ID: "commitlint"}},
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
			var refErr *ReferenceError
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
		cfg       Config
		wantCat   string
		wantKey   string
		wantField string
		wantRef   string // "" means Validate must return nil
	}{
		{
			name: "tool runtime, missing",
			cfg: Config{
				Tools:    map[string]Tool{"eslint": {Name: "eslint", Runtime: "node"}},
				Runtimes: CategoryConfig[Runtime]{Definitions: map[string]Runtime{}},
			},
			wantCat: "tool", wantKey: "eslint", wantField: "runtime", wantRef: "node",
		},
		{
			name: "tool runtime, present",
			cfg: Config{
				Tools:    map[string]Tool{"eslint": {Name: "eslint", Runtime: "node"}},
				Runtimes: CategoryConfig[Runtime]{Definitions: map[string]Runtime{"node": {Type: "node"}}},
			},
		},
		{
			name: "tool download, missing",
			cfg: Config{
				Tools:     map[string]Tool{"shellcheck": {Name: "shellcheck", Download: "shellcheck"}},
				Downloads: map[string]Download{},
			},
			wantCat: "tool", wantKey: "shellcheck", wantField: "download", wantRef: "shellcheck",
		},
		{
			name: "tool download, present",
			cfg: Config{
				Tools:     map[string]Tool{"shellcheck": {Name: "shellcheck", Download: "shellcheck"}},
				Downloads: map[string]Download{"shellcheck": {Name: "shellcheck"}},
			},
		},
		{
			name: "tool with neither runtime nor download set",
			cfg: Config{
				Tools: map[string]Tool{"standalone": {Name: "standalone"}},
			},
		},
		{
			name: "runtime download, missing",
			cfg: Config{
				Runtimes:  CategoryConfig[Runtime]{Definitions: map[string]Runtime{"node": {Type: "node", Download: "node"}}},
				Downloads: map[string]Download{},
			},
			wantCat: "runtime", wantKey: "node", wantField: "download", wantRef: "node",
		},
		{
			name: "runtime download, present",
			cfg: Config{
				Runtimes:  CategoryConfig[Runtime]{Definitions: map[string]Runtime{"node": {Type: "node", Download: "node"}}},
				Downloads: map[string]Download{"node": {Name: "node"}},
			},
		},
		{
			name: "lint tools, missing",
			cfg: Config{
				Lint:  CategoryConfig[Linter]{Definitions: map[string]Linter{"eslint": {Name: "eslint", Tools: []string{"eslint-bin"}}}},
				Tools: map[string]Tool{},
			},
			wantCat: "lint", wantKey: "eslint", wantField: "tools", wantRef: "eslint-bin",
		},
		{
			name: "lint tools, present",
			cfg: Config{
				Lint:  CategoryConfig[Linter]{Definitions: map[string]Linter{"eslint": {Name: "eslint", Tools: []string{"eslint-bin"}}}},
				Tools: map[string]Tool{"eslint-bin": {Name: "eslint-bin"}},
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
			var refErr *ReferenceError
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
	cfg := Config{
		Runtimes: CategoryConfig[Runtime]{Enabled: []string{"node"}, Definitions: map[string]Runtime{}},
		Lint:     CategoryConfig[Linter]{Enabled: []string{"eslint"}, Definitions: map[string]Linter{}},
	}

	err := cfg.Validate()
	require.Error(t, err)

	joined, ok := err.(interface{ Unwrap() []error })
	require.True(t, ok, "errors.Join must return something implementing Unwrap() []error")
	assert.Len(t, joined.Unwrap(), 2)
}
