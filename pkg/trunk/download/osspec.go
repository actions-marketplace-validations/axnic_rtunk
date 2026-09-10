package download

import "github.com/xunleii/rtunk/pkg/trunk/config"

// goosNames maps Go's GOOS to trunk's own os vocabulary (ARCHITECTURE.md "downloads[].os"),
// the key a DownloadEntry.OS map is keyed by.
var goosNames = map[string]string{"darwin": "macos", "linux": "linux", "windows": "windows"}

// goarchNames is goosNames' CPU equivalent.
var goarchNames = map[string]string{"amd64": "x86_64", "arm64": "arm_64"}

// MatchEntry returns the first entries[] whose OS/CPU both cover goos/goarch, along with the
// upstream-naming values (DownloadEntry.OS/CPU's map values) to template into ${os}/${cpu}. ok is
// false if nothing matches -- no entry for this platform, or Go's GOOS/GOARCH has no mapping in
// trunk's vocabulary.
//
// ponytail: first match wins; DownloadEntry.Version range-gated entries (e.g. hadolint's
// "<2.13.1" vs ">=2.13.1" URL schemes, ARCHITECTURE.md "downloads[].version") aren't
// disambiguated by the version being fetched -- add semver-range comparison here if a plugin
// this actually matters for comes up.
func MatchEntry(entries []config.DownloadEntry, goos, goarch string) (config.DownloadEntry, string, string, bool) {
	osName, ok := goosNames[goos]
	if !ok {
		return config.DownloadEntry{}, "", "", false
	}
	cpuName, ok := goarchNames[goarch]
	if !ok {
		return config.DownloadEntry{}, "", "", false
	}
	for _, e := range entries {
		osVal, ok := e.OS[osName]
		if !ok {
			continue
		}
		cpuVal, ok := e.CPU[cpuName]
		if !ok {
			continue
		}
		return e, osVal, cpuVal, true
	}
	return config.DownloadEntry{}, "", "", false
}
