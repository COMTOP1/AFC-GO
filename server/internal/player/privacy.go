package player

import (
	"time"

	"github.com/COMTOP1/AFC-GO/server/internal/role"
)

// adultAge is the age from which a player's photo may be shown.
const adultAge = 18

// Age returns the age in whole years on now's date; ok is false when dob is
// after today (bad data).
func Age(dob, now time.Time) (int, bool) {
	now = now.In(dob.Location())
	ty, tm, td := now.Date()
	today := time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := dob.Date()
	birth := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birth) {
		return 0, false
	}
	age := ty - by
	if birth.AddDate(age, 0, 0).After(today) {
		age--
	}
	return age, true
}

// PhotoVisible is the single safeguarding rule for player photos: never for
// players on a youth team, never for players under 18, and never when the
// date of birth is nonsensical. Players with no date of birth on a senior
// team are shown, matching the legacy site.
func PhotoVisible(p Player, teamIsYouth bool, now time.Time) bool {
	if !p.FileName.Valid || p.FileName.String == "" || teamIsYouth {
		return false
	}
	if !p.DateOfBirth.Valid {
		return true
	}
	age, ok := Age(p.DateOfBirth.Time, now)
	return ok && age >= adultAge
}

// DateOfBirthVisible is the rule for who sees a player's date of birth (and
// age) in the players list: anyone who can edit players, since they enter
// it, and a Manager for players on their own team. Everyone else, such as a
// Photographer or another team's Manager, has no need for it.
// An unknown or empty role sees nothing: CanEdit alone would let it through.
func DateOfBirthVisible(viewer role.Role, viewerTeamID, playerTeamID int) bool {
	if viewer.DBString() == "" {
		return false
	}
	if viewer.CanEdit() {
		return true
	}
	return viewer == role.Manager && viewerTeamID != 0 && viewerTeamID == playerTeamID
}
