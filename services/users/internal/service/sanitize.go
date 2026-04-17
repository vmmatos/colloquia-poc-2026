package service

import "strings"

// EscapeLikePattern escapes PostgreSQL LIKE/ILIKE metacharacters in s so that
// the value is treated as a literal string in a LIKE/ILIKE pattern.
// The backslash is PostgreSQL's default LIKE escape character.
func EscapeLikePattern(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}
