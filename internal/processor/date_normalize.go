package processor

import (
	"fmt"
	"strings"
	"time"
)

const (
	dateLayoutDash = "2006-01-02"
	dateLayoutSlash = "2006/01/02"
	dateLayoutCompact = "20060102"
)

func NormalizeDate(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("日期为空")
	}

	if t, ok := parseByLayouts(s, []string{dateLayoutDash, dateLayoutSlash, dateLayoutCompact}); ok {
		return t.Format(dateLayoutDash), nil
	}
	return "", fmt.Errorf("无法解析日期: %q", input)
}

func parseByLayouts(s string, layouts []string) (time.Time, bool) {
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

