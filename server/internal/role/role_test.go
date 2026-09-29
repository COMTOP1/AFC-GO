package role_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
)

func TestPredicates(t *testing.T) {
	all := []role.Role{role.Photographer, role.Manager, role.ProgrammeEditor, role.LeagueSecretary,
		role.Treasurer, role.SafeguardingOfficer, role.ClubSecretary, role.Chairperson, role.Webmaster}
	for _, r := range all {
		assert.Equal(t, r != role.Manager && r != role.Photographer, r.CanEdit(), "CanEdit %s", r)
		assert.Equal(t, r != role.Manager, r.CanManageGallery(), "CanManageGallery %s", r)
		want := r == role.SafeguardingOfficer || r == role.ClubSecretary || r == role.Chairperson || r == role.Webmaster
		assert.Equal(t, want, r.IsClubSecretaryHigher(), "IsClubSecretaryHigher %s", r)
	}
}
