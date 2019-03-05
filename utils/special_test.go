package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecialTransform(t *testing.T) {
	a := "[Get5] Match is LIVE"
	b := "[\x05Get5\x01] Match is \x04LIVE\n"

	c := StripSpecial(b)

	assert.Equal(t, a, c)
}
