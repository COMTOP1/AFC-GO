package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/auth"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

var ctx = context.Background()

// fakeUsers accepts password "Right-Pa55" for known emails; reset-flagged
// accounts only report reset-required after a correct password (Task 21).
type fakeUsers struct {
	byEmail map[string]user.User
	changed map[int]string
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byEmail: map[string]user.User{
			"ok@example.test":    {ID: 1, Email: "ok@example.test", Role: role.Treasurer},
			"reset@example.test": {ID: 2, Email: "reset@example.test", Role: role.Treasurer, ResetPassword: true},
		},
		changed: map[int]string{},
	}
}

func (f *fakeUsers) GetUser(_ context.Context, u user.User) (user.User, error) {
	for _, x := range f.byEmail {
		if x.ID == u.ID {
			return x, nil
		}
	}
	return user.User{}, fmt.Errorf("failed to get user: %w", sql.ErrNoRows)
}

func (f *fakeUsers) VerifyUser(_ context.Context, u user.User, _, _, _, _, _ int) (user.User, bool, error) {
	x, ok := f.byEmail[u.Email]
	if !ok || u.Password.String != "Right-Pa55" {
		return u, false, errors.New("invalid credentials")
	}
	if x.ResetPassword {
		return x, true, errors.New("password reset required")
	}
	return x, false, nil
}

func (f *fakeUsers) EditUserPassword(_ context.Context, u user.User, _, _, _, _ int) error {
	f.changed[u.ID] = u.Password.String
	return nil
}

func newAuthService() (*auth.Service, *fakeUsers, *auth.Tokens) {
	users := newFakeUsers()
	tokens := auth.NewTokens(auth.RedisConfig{})
	return auth.NewService(users, tokens, auth.PasswordConfig{}), users, tokens
}

func TestPasswordProblems(t *testing.T) {
	assert.Empty(t, auth.PasswordProblems("Str0ng-Pass!"))
	for _, weak := range []string{"short1!A", "alllowercase1!", "ALLUPPERCASE1!", "NoDigits!!aa", "NoSpecial1aaA"} {
		assert.NotEmpty(t, auth.PasswordProblems(weak), weak)
	}
}

func TestLogin(t *testing.T) {
	svc, _, tokens := newAuthService()

	res, err := svc.Login(ctx, "ok@example.test", "Right-Pa55")
	require.NoError(t, err)
	assert.Equal(t, 1, res.User.ID)
	assert.False(t, res.ResetRequired)

	_, err = svc.Login(ctx, "ok@example.test", "wrong")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
	_, err = svc.Login(ctx, "reset@example.test", "wrong")
	require.ErrorIs(t, err, auth.ErrInvalidCredentials, "no reset link without the password")

	res, err = svc.Login(ctx, "reset@example.test", "Right-Pa55")
	require.NoError(t, err)
	assert.True(t, res.ResetRequired)
	require.True(t, strings.HasPrefix(res.ResetURL, "/reset/"))
	id, ok := tokens.Get(ctx, strings.TrimPrefix(res.ResetURL, "/reset/"))
	assert.True(t, ok)
	assert.Equal(t, 2, id)
}

func TestChangePassword(t *testing.T) {
	svc, users, _ := newAuthService()
	me := users.byEmail["ok@example.test"]
	fields := func(err error) map[string]string {
		se, ok := svcerr.As(err)
		require.True(t, ok, "want svcerr, got %v", err)
		return se.Fields
	}
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "wrong", "Str0ng-Pass!", "Str0ng-Pass!")), "oldPassword")
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "Right-Pa55", "Str0ng-Pass!", "different")), "confirmationPassword")
	assert.Contains(t, fields(svc.ChangePassword(ctx, me, "Right-Pa55", "weak", "weak")), "newPassword")
	require.NoError(t, svc.ChangePassword(ctx, me, "Right-Pa55", "Str0ng-Pass!", "Str0ng-Pass!"))
	assert.Equal(t, "Str0ng-Pass!", users.changed[1])
}

func TestResetFlow(t *testing.T) {
	svc, users, tokens := newAuthService()
	require.NoError(t, tokens.Set(ctx, "tok", 2, time.Hour))

	require.NoError(t, svc.CheckResetToken(ctx, "tok"))
	se, ok := svcerr.As(svc.CheckResetToken(ctx, "nope"))
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)

	se, ok = svcerr.As(svc.ResetPassword(ctx, "tok", "Str0ng-Pass!", "mismatch"))
	require.True(t, ok)
	assert.Contains(t, se.Fields, "confirmationPassword")

	require.NoError(t, svc.ResetPassword(ctx, "tok", "Str0ng-Pass!", "Str0ng-Pass!"))
	assert.Equal(t, "Str0ng-Pass!", users.changed[2])
	_, still := tokens.Get(ctx, "tok")
	assert.False(t, still, "tokens are single use")
}
