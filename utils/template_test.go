package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
	variables := map[string]string{
		"a":  "b",
		"cc": "d",
	}
	actual := RenderTemplate("${a}..${cc}..${e}", variables)
	assert.Equal(t, "b..d..${e}", actual)
}
