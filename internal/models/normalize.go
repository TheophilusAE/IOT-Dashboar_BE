package models

import "strings"

func normalizeToken(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
