package util

import (
	"strings"
)

func JoinObjectName(array ...string) string {
	vec := make([]string, 0, len(array))
	for _, s := range array {
		if s == "" {
			continue
		}
		vec = append(vec, s)
	}
	return strings.ToLower(strings.Join(vec, "."))
}
