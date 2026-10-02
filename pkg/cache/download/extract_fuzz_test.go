package download_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/axnic/rtunk/pkg/cache/download"
	"github.com/axnic/rtunk/pkg/trunk/config"
)

// FuzzInstallDownloadTarGz feeds archives with arbitrary entry names and symlink targets to the
// extractor. The archive comes from the network, so whatever it contains, nothing may end up
// outside the destination directory (zip slip, CWE-22, and symlink chains, CWE-59). Extraction
// may fail; it may not write elsewhere. The seed corpus runs on every `go test`;
// `go test -fuzz=FuzzInstallDownloadTarGz ./pkg/cache/download` explores beyond it.
func FuzzInstallDownloadTarGz(f *testing.F) {
	f.Add("bin", "x", "lib", "y", "bin/tool")
	f.Add("a", "../..", "b", "../../..", "../escape")
	f.Add("link", "/etc/passwd", "l2", "link", "l2/x")
	// The symlink chain of TestInstallDownload_TarGz_SymlinkChainEscape: each entry looks harmless
	// on its own, together they climb out of the destination.
	f.Add("s", ".", "s/s/s/up", "../..", "s/s/s/up/pwned.txt")
	f.Fuzz(func(t *testing.T, name1, target1, name2, target2, name3 string) {
		dir := t.TempDir()
		blob := filepath.Join(dir, "blob")

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		// Two symlinks, then a regular file that may be written through whatever they created.
		// "top/" is stripped by StripComponents.
		for _, l := range [][2]string{{name1, target1}, {name2, target2}} {
			_ = tw.WriteHeader(&tar.Header{Name: "top/" + l[0], Typeflag: tar.TypeSymlink, Linkname: l[1]})
		}
		if tw.WriteHeader(&tar.Header{Name: "top/" + name3, Mode: 0o644, Size: 1}) == nil {
			_, _ = tw.Write([]byte("x"))
		}
		_ = tw.Close()
		_ = gz.Close()
		if err := os.WriteFile(blob, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}

		// dest sits two levels below dir: anything written above it, through ".." or a symlink, lands
		// in dir or dir/a, inside t.TempDir().
		dest := filepath.Join(dir, "a", "install")
		_ = download.InstallDownload(blob, "https://example.com/archive.tar.gz", dest, config.DownloadEntry{StripComponents: 1}, "")

		top, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range top {
			if e.Name() != "blob" && e.Name() != "a" {
				t.Fatalf("extraction wrote %q outside the destination", filepath.Join(dir, e.Name()))
			}
		}
		if inner, err := os.ReadDir(filepath.Join(dir, "a")); err == nil {
			for _, e := range inner {
				if e.Name() != "install" {
					t.Fatalf("extraction wrote %q next to the destination", filepath.Join(dir, "a", e.Name()))
				}
			}
		}
	})
}
