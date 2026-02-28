package processor

import (
	"strings"
)

var windowsInvalidFileNameChars = []string{
	"\\", "/", ":", "*", "?", "\"", "<", ">", "|",
}

const (
	replaceChar = "-"
)

func SanitizeFileNamePart(input string) string {
	s := strings.TrimSpace(input)
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for _, ch := range windowsInvalidFileNameChars {
		s = strings.ReplaceAll(s, ch, replaceChar)
	}
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, ". ")
	return s
}

