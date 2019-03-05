package utils

import (
	"golang.org/x/text/transform"
)

var stripSpecialTransformer = transform.RemoveFunc(func(r rune) bool {
	return r < 32
})

// StripSpecial will remove all special, unreadable chars
func StripSpecial(line string) string {
	line, _, err := transform.String(stripSpecialTransformer, line)
	if err != nil {
		panic(err)
	}
	return line
}
