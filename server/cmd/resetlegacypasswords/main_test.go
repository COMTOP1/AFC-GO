package main

import (
	"context"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

type fakeResetter struct {
	reset []int
	fail  map[int]bool
}

func (f *fakeResetter) ResetPassword(_ context.Context, id int) (user.ResetResult, error) {
	f.reset = append(f.reset, id)
	if f.fail[id] {
		return user.ResetResult{}, errors.New("smtp down")
	}
	return user.ResetResult{EmailSent: true}, nil
}

func addLegacy(t *testing.T, db *sqlx.DB, email string) int {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `ALTER TABLE users ADD COLUMN IF NOT EXISTS password TEXT`)
	require.NoError(t, err)
	var id int
	require.NoError(t, db.GetContext(context.Background(), &id,
		`INSERT INTO users (name, email, role, password) VALUES ('Legacy', $1, 'TREASURER', 'digest') RETURNING id`, email))
	return id
}

func TestResetAllEmailsEveryClearedUser(t *testing.T) {
	ctx := context.Background()
	db, _ := testdb.Open(t)
	first := addLegacy(t, db, "a@example.test")
	second := addLegacy(t, db, "b@example.test")

	r := &fakeResetter{fail: map[int]bool{second: true}}
	assert.Equal(t, 1, resetAll(ctx, user.NewUserRepo(db), r), "a failed email is reported")
	assert.ElementsMatch(t, []int{first, second}, r.reset)

	r = &fakeResetter{}
	assert.Equal(t, 0, resetAll(ctx, user.NewUserRepo(db), r))
	assert.Empty(t, r.reset, "already-cleared accounts aren't emailed again")
}
