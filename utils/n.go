package utils

import "strings"

func Normalize(s string) string {
	return strings.Title(strings.ToLower(strings.TrimSpace(s)))
}
