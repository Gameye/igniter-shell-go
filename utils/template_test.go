package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
	variables := map[string]string{
		"a": "b",
		"c": "d",
	}
	actual := RenderTemplate("${a}..${c}..${e}", variables)
	assert.Equal(t, "b..d..${e}", actual)
}
