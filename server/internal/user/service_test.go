package user_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/upload/uploadtest"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var ctx = context.Background()

func secretary() user.User {
	return user.User{ID: 3, Name: "Club Secretary", Email: "secretary@example.test", Phone: null.StringFrom("0123"),
		Role: role.ClubSecretary, FileName: null.StringFrom("user/sec.png"),
		Hash: null.StringFrom("keep-me"), Salt: null.StringFrom("keep-me-too")}
}

type harness struct {
	svc     *user.Service
	store   *fakeStore
	objects *uploadtest.Storage
	sender  *fakeSender
	tokens  *fakeTokens
}

func newHarness(senderErr error) harness {
	h := harness{
		store:   newFakeStore(secretary()),
		objects: uploadtest.New(),
		sender:  &fakeSender{err: senderErr},
		tokens:  &fakeTokens{tokens: map[string]int{}},
	}
	h.objects.Objects["user/sec.png"] = "IMG"
	teams := fakeTeams{1: team.Team{ID: 1, Name: "First Team"}}
	h.svc = user.NewService(h.store, teams, upload.New(h.objects), h.sender, h.tokens,
		user.HashParams{WorkFactor: 2, BlockSize: 1, Parallelism: 1, KeyLength: 32}, "afc.example.test")
	return h
}

func fieldsOf(t *testing.T, err error) map[string]string {
	t.Helper()
	se, ok := svcerr.As(err)
	require.True(t, ok, "want svcerr, got %v", err)
	return se.Fields
}

func TestCreateValidates(t *testing.T) {
	h := newHarness(nil)
	for field, in := range map[string]user.CreateInput{
		"name":   {Email: "a@b.test", Role: "treasurer"},
		"email":  {Name: "x", Email: "not-email", Role: "treasurer"},
		"role":   {Name: "x", Email: "a@b.test", Role: "president"},
		"teamId": {Name: "x", Email: "a@b.test", Role: "manager", TeamID: 9},
	} {
		_, err := h.svc.Create(ctx, in, nil)
		assert.Contains(t, fieldsOf(t, err), field)
	}
}

func TestCreateSendsSignupEmail(t *testing.T) {
	h := newHarness(nil)
	created, err := h.svc.Create(ctx, user.CreateInput{Name: "New", Email: "new@example.test", Role: "manager", TeamID: 1}, nil)
	require.NoError(t, err)
	assert.True(t, created.EmailSent)
	assert.Empty(t, created.TempPassword, "never return the password when it was emailed")
	require.Len(t, h.sender.sent, 1)
	assert.Equal(t, "new@example.test", h.sender.sent[0].To)

	row := h.store.row(created.User.ID)
	assert.True(t, row.ResetPassword, "new users must reset their password on first login")
	assert.True(t, row.Hash.Valid)
	assert.Equal(t, 1, row.TeamID)
}

func TestCreateWithoutMailerReturnsPassword(t *testing.T) {
	h := newHarness(emails.ErrNoMailer)
	created, err := h.svc.Create(ctx, user.CreateInput{Name: "New", Email: "new@example.test", Role: "treasurer", TeamID: 1}, nil)
	require.NoError(t, err)
	assert.False(t, created.EmailSent)
	assert.NotEmpty(t, created.TempPassword)
	assert.Zero(t, h.store.row(created.User.ID).TeamID, "only managers keep a team")
}

func TestCreateDuplicateEmailIsConflict(t *testing.T) {
	h := newHarness(nil)
	_, err := h.svc.Create(ctx, user.CreateInput{Name: "Dup", Email: "secretary@example.test", Role: "treasurer"}, nil)
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindConflict, se.Kind)
}

func TestUpdatePreservesPasswordAndClearsPhone(t *testing.T) {
	h := newHarness(nil)
	empty := ""
	_, err := h.svc.Update(ctx, 3, user.UpdateInput{Phone: &empty}, nil)
	require.NoError(t, err)
	row := h.store.row(3)
	assert.False(t, row.Phone.Valid)
	assert.Equal(t, "keep-me", row.Hash.String, "profile edits never touch credentials")
	assert.Equal(t, role.ClubSecretary, row.Role)
}

func TestDelete(t *testing.T) {
	h := newHarness(nil)
	deleted, err := h.svc.Delete(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, "Club Secretary", deleted.Name)
	assert.Equal(t, []string{"user/sec.png"}, h.objects.Deleted)
}

func TestResetPassword(t *testing.T) {
	h := newHarness(nil)
	res, err := h.svc.ResetPassword(ctx, 3)
	require.NoError(t, err)
	assert.True(t, res.EmailSent)
	assert.Empty(t, res.ResetURL)
	assert.True(t, h.store.row(3).ResetPassword)
	require.Len(t, h.tokens.tokens, 1)
	assert.Equal(t, 7*24*time.Hour, h.tokens.ttl)
	for token, id := range h.tokens.tokens {
		assert.Equal(t, 3, id)
		assert.Contains(t, fmt.Sprint(h.sender.sent[0].TplData), token)
	}

	h = newHarness(emails.ErrNoMailer)
	res, err = h.svc.ResetPassword(ctx, 3)
	require.NoError(t, err)
	assert.False(t, res.EmailSent)
	assert.True(t, strings.HasPrefix(res.ResetURL, "https://afc.example.test/reset/"))
}
