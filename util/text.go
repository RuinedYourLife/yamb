package util

import "regexp"

func StripANSICodes(s string) string {
	re := regexp.MustCompile(`\x1B\[[0-?]*[ -/]*[@-~]`)
	return re.ReplaceAllString(s, "")
}
