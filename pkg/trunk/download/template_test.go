package download_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xunleii/rtunk/pkg/trunk/download"
)

func TestTemplateURL(t *testing.T) {
	got := download.TemplateURL(
		"https://github.com/koalaman/shellcheck/releases/download/v${version}/shellcheck-v${version}.${os}.${cpu}.tar.xz",
		"0.11.0", "linux", "x86_64",
	)
	want := "https://github.com/koalaman/shellcheck/releases/download/v0.11.0/shellcheck-v0.11.0.linux.x86_64.tar.xz"
	assert.Equal(t, want, got)
}
