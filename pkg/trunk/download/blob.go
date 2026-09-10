package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// ProgressFunc reports bytes received so far out of total (total is -1 if the server didn't send
// Content-Length). Called from FetchBlob's own goroutine; callers that forward it onto a channel
// (see Event) are responsible for their own synchronization.
type ProgressFunc func(bytes, total int64)

// FetchBlob downloads url over HTTPS (or HTTP, for tests) into root's content-addressed blob
// store: streamed through a SHA256 hasher into a temp file, renamed to BlobPath(root, sum) only
// once the hash is known, so a path is never read unless its name matches its own content (see
// the spec's "Checksum model"). progress may be nil.
func FetchBlob(root, url string, progress ProgressFunc) (string, error) {
	resp, err := http.Get(url) //nolint:noctx // v0.2 has no per-fetch context/cancellation yet
	if err != nil {
		return "", fmt.Errorf("download: fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download: fetch %s: unexpected status %s", url, resp.Status)
	}

	blobsDir := filepath.Join(root, "blobs", "sha256")
	if err := os.MkdirAll(blobsDir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(blobsDir, ".tmp-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed below

	hasher := sha256.New()
	var written int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := tmp.Write(buf[:n]); err != nil {
				tmp.Close()
				return "", err
			}
			hasher.Write(buf[:n])
			written += int64(n)
			if progress != nil {
				progress(written, resp.ContentLength)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			tmp.Close()
			return "", fmt.Errorf("download: fetch %s: %w", url, readErr)
		}
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	dst := BlobPath(root, sum)
	if err := os.Rename(tmp.Name(), dst); err != nil {
		return "", err
	}
	return dst, nil
}
