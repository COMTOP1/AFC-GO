package sponsor_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/sponsor"
)

// TestPublicTeamOmitsEmpty pins the MINOR fix: the store's minimal/team
// queries don't select team_id, so Team is always "" on those rows. Without
// omitempty, that emitted a misleading "team":"" instead of leaving the
// field out like every other optional field on Public.
func TestPublicTeamOmitsEmpty(t *testing.T) {
	b, err := json.Marshal(sponsor.Public{ID: 1, Name: "Club Sponsor"})
	require.NoError(t, err)
	assert.NotContains(t, string(b), `"team"`, "empty team must be omitted, got %s", b)

	b, err = json.Marshal(sponsor.Public{ID: 2, Name: "Team Sponsor", Team: "1"})
	require.NoError(t, err)
	assert.Contains(t, string(b), `"team":"1"`, "a set team must still be serialised, got %s", b)
}
