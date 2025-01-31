package internal

import (
	"net/url"
	"strings"
)

func PathEscape(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "/", "%2F")
}
