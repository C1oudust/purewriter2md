package main

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var invalidPathChars = regexp.MustCompile(`[\\/:*?"<>|]`)

func ParseTime(t int64, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Unix(t/1000, 0).Format(format)
}

func SanitizePathName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Untitled"
	}
	name = invalidPathChars.ReplaceAllString(name, "x")
	name = strings.TrimSpace(name)
	if name == "" {
		return "Untitled"
	}
	return name
}

func BuildUntitledFromContent(content string, maxRunes int) string {
	text := strings.TrimSpace(content)
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}
	if utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	runes := []rune(text)
	return strings.TrimSpace(string(runes[:maxRunes]))
}
