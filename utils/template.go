package utils

import "regexp"

var re = regexp.MustCompile(`\$\{([\w\.]+)\}`)

// RenderTemplate renders a template!
func RenderTemplate(
	template string,
	variables map[string]string,
) string {
	return re.ReplaceAllStringFunc(
		template,
		func(s string) string {
			m := re.FindStringSubmatch(s)
			r, ok := variables[m[1]]
			if ok {
				return r
			}
			return m[0]
		},
	)
}
