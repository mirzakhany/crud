package admin

import (
	"strings"

	"golang.org/x/text/cases"
)

func replace(old, new, src string) string {
	return strings.Replace(src, old, new, -1)
}

func title(str string) string {
	return cases.Title().String(str)
}
