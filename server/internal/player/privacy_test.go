package player_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/player"
)

var now = time.Date(2026, 9, 27, 15, 0, 0, 0, time.UTC)

func born(y, m, d int) null.Time {
	return null.TimeFrom(time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC))
}

func TestAge(t *testing.T) {
	cases := []struct {
		dob  null.Time
		want int
		ok   bool
	}{
		{born(2008, 9, 27), 18, true}, // 18th birthday today
		{born(2008, 9, 28), 17, true}, // tomorrow
		{born(1990, 1, 1), 36, true},
		{born(2030, 1, 1), 0, false}, // future
	}
	for _, tc := range cases {
		got, ok := player.Age(tc.dob.Time, now)
		assert.Equal(t, tc.ok, ok, tc.dob.Time)
		if tc.ok {
			assert.Equal(t, tc.want, got, tc.dob.Time)
		}
	}
}

// TestPhotoVisible pins Review Focus #1.
func TestPhotoVisible(t *testing.T) {
	photo := null.StringFrom("player/p.png")
	cases := map[string]struct {
		p     player.Player
		youth bool
		want  bool
	}{
		"adult, senior team":           {player.Player{FileName: photo, DateOfBirth: born(1990, 1, 1)}, false, true},
		"adult, youth team":            {player.Player{FileName: photo, DateOfBirth: born(1990, 1, 1)}, true, false},
		"17, senior team":              {player.Player{FileName: photo, DateOfBirth: born(2008, 9, 28)}, false, false},
		"18 today, senior team":        {player.Player{FileName: photo, DateOfBirth: born(2008, 9, 27)}, false, true},
		"no DOB, senior team":          {player.Player{FileName: photo}, false, true},
		"no DOB, youth team":           {player.Player{FileName: photo}, true, false},
		"future DOB (bad data)":        {player.Player{FileName: photo, DateOfBirth: born(2030, 1, 1)}, false, false},
		"no photo":                     {player.Player{DateOfBirth: born(1990, 1, 1)}, false, false},
		"empty-but-valid photo string": {player.Player{FileName: null.StringFrom(""), DateOfBirth: born(1990, 1, 1)}, false, false},
	}
	for name, tc := range cases {
		assert.Equal(t, tc.want, player.PhotoVisible(tc.p, tc.youth, now), name)
	}
}
