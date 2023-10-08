package crud

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func replace(old, new, src string) string {
	return strings.Replace(src, old, new, -1)
}

func title(str string) string {
	return cases.Title(language.English).String(str)
}

func lower(str string) string {
	return cases.Lower(language.English).String(str)
}

func upper(str string) string {
	return cases.Upper(language.English).String(str)
}

func templateFuncs() map[string]any {
	return map[string]any{
		"replace": replace,
		"title":   title,
		"lower":   lower,
		"upper":   upper,
	}
}
