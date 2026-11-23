package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

func TestIsWebURL(t *testing.T) {
	for _, ok := range []string{
		"https://example.com",
		"http://example.com/table?x=1",
		"HTTPS://Example.com/path",
	} {
		assert.True(t, utils.IsWebURL(ok), ok)
	}
	for _, bad := range []string{
		"",
		"javascript:alert(1)",
		"JavaScript://example.com/%0aalert(1)",
		"data:text/html,<script>alert(1)</script>",
		"vbscript:msgbox(1)",
		"mailto:a@example.com",
		"ftp://example.com",
		"https://",
		"https:///path",
		"/relative/path",
		"example.com",
	} {
		assert.False(t, utils.IsWebURL(bad), bad)
	}
}
