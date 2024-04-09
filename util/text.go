package util

import (
	"regexp"
	"strings"
)

func StripANSICodes(s string) string {
	re := regexp.MustCompile(`\x1B\[[0-?]*[ -/]*[@-~]`)
	return re.ReplaceAllString(s, "")
}

func SanitizeLowerString(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	return strings.ToLower(re.ReplaceAllString(s, ""))
}
