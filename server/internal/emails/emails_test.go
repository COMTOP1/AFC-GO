package emails_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
)

func TestSignupRendersCredentials(t *testing.T) {
	m, err := emails.Signup("new@example.test", "New Person", "Tmp-Pa55!", "afc.example.test")
	require.NoError(t, err)
	assert.Equal(t, "new@example.test", m.To)
	var body bytes.Buffer
	require.NoError(t, m.Tpl.Execute(&body, m.TplData))
	for _, want := range []string{"New Person", "new@example.test", "Tmp-Pa55!", "https://afc.example.test", "https://afc.example.test/AFC.png", `lang="en-gb"`} {
		assert.Contains(t, body.String(), want)
	}
	assert.NotContains(t, body.String(), "github.com/COMTOP1")
}

func TestResetRendersLink(t *testing.T) {
	m, err := emails.Reset("user@example.test", "https://afc.example.test/reset/abc", "afc.example.test")
	require.NoError(t, err)
	var body bytes.Buffer
	require.NoError(t, m.Tpl.Execute(&body, m.TplData))
	assert.Contains(t, body.String(), "https://afc.example.test/reset/abc")
	assert.Contains(t, body.String(), `lang="en-gb"`)
	assert.Contains(t, body.String(), "valid for one week as of the sent time")
	assert.Contains(t, body.String(), "https://afc.example.test/AFC.png")
	assert.NotContains(t, body.String(), "github.com/COMTOP1")
}
