package download

import "strings"

// TemplateURL substitutes a download recipe's ${version}/${os}/${cpu} placeholders
// (ARCHITECTURE.md "downloads: url") with resolved values.
func TemplateURL(url, version, os, cpu string) string {
	r := strings.NewReplacer("${version}", version, "${os}", os, "${cpu}", cpu)
	return r.Replace(url)
}
