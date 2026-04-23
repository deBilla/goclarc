// Package gen provides code generation utilities.
package gen

import (
	"strings"
	"unicode"
)

// initialisms is the set of well-known Go initialisms that must stay uppercase.
// Matches the list used by the Go tools (golint, staticcheck).
var initialisms = map[string]bool{
	"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true,
	"DNS": true, "EOF": true, "GUID": true, "HTML": true, "HTTP": true,
	"HTTPS": true, "ID": true, "IP": true, "JSON": true, "LHS": true,
	"QPS": true, "RAM": true, "RHS": true, "RPC": true, "SLA": true,
	"SMTP": true, "SQL": true, "SSH": true, "TCP": true, "TLS": true,
	"TTL": true, "UDP": true, "UI": true, "UID": true, "UUID": true,
	"URI": true, "URL": true, "UTF8": true, "VM": true, "XML": true,
	"XMPP": true, "XSRF": true, "XSS": true,
}

// splitWords splits a snake_case or camelCase string into its component words.
func splitWords(s string) []string {
	// Replace hyphens and dots with underscores for uniform splitting.
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, ".", "_")

	if strings.Contains(s, "_") {
		parts := strings.Split(s, "_")
		var words []string
		for _, p := range parts {
			if p != "" {
				words = append(words, p)
			}
		}
		return words
	}

	// camelCase split.
	var words []string
	var current strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// capitalise returns the word with its first letter uppercased.
// If the uppercased word is a known initialism, it returns the full initialism.
func capitalise(word string) string {
	if word == "" {
		return ""
	}
	upper := strings.ToUpper(word)
	if initialisms[upper] {
		return upper
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// ToPascal converts snake_case or camelCase to PascalCase respecting Go initialisms.
// Examples: "user_id" → "UserID", "video_card" → "VideoCard", "http_url" → "HTTPURL"
func ToPascal(s string) string {
	words := splitWords(s)
	var b strings.Builder
	for _, w := range words {
		b.WriteString(capitalise(w))
	}
	return b.String()
}

// ToCamel converts snake_case to camelCase respecting Go initialisms.
// Examples: "user_id" → "userID", "video_card" → "videoCard"
func ToCamel(s string) string {
	words := splitWords(s)
	if len(words) == 0 {
		return s
	}
	var b strings.Builder
	b.WriteString(strings.ToLower(words[0]))
	for _, w := range words[1:] {
		b.WriteString(capitalise(w))
	}
	return b.String()
}

// ToKebab converts snake_case or PascalCase to kebab-case.
// Examples: "videoCard" → "video-card", "user_id" → "user-id"
func ToKebab(s string) string {
	words := splitWords(s)
	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}
	return strings.Join(lower, "-")
}

// ToSnake converts PascalCase or camelCase to snake_case.
// Examples: "UserID" → "user_id", "VideoCard" → "video_card"
func ToSnake(s string) string {
	words := splitWords(s)
	lower := make([]string, len(words))
	for i, w := range words {
		lower[i] = strings.ToLower(w)
	}
	return strings.Join(lower, "_")
}

// ToPlural returns a naive plural of a word (appends "s", handles common endings).
func ToPlural(s string) string {
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "z") || strings.HasSuffix(s, "ch") ||
		strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 {
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

// BaseType strips a leading "*" from a type string.
// Example: "*time.Time" → "time.Time", "string" → "string"
func BaseType(s string) string {
	return strings.TrimPrefix(s, "*")
}
