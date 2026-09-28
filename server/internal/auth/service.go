package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/user"
)

// ErrInvalidCredentials is returned for any failed login, without saying why.
var ErrInvalidCredentials = errors.New("invalid email or password")

const loginResetTTL = time.Hour

// UserStore is the credential side of the user store (satisfied by *user.Store).
type UserStore interface {
	GetUser(ctx context.Context, userParam user.User) (user.User, error)
	VerifyUser(ctx context.Context, userParam user.User, iter, workFactor, blockSize, parallelismFactor, keyLen int) (user.User, bool, error)
	EditUserPassword(ctx context.Context, userParam user.User, workFactor, blockSize, parallelismFactor, keyLen int) error
}

// LoginResult is a successful credential check. When ResetRequired is set
// no session should be created; the user must visit ResetURL.
type LoginResult struct {
	User          user.User
	ResetRequired bool
	ResetURL      string
}

// Service handles logins and password changes.
type Service struct {
	users     UserStore
	tokens    *Tokens
	passwords PasswordConfig
}

func NewService(users UserStore, tokens *Tokens, passwords PasswordConfig) *Service {
	return &Service{users: users, tokens: tokens, passwords: passwords}
}

var passwordRules = []struct {
	re  *regexp.Regexp
	msg string
}{
	{regexp.MustCompile(`[a-z]`), "at least 1 lower case letter"},
	{regexp.MustCompile(`[A-Z]`), "at least 1 upper case letter"},
	{regexp.MustCompile(`\d`), "at least 1 number"},
	{regexp.MustCompile(`[@$!%*?&|^£;:/.,<>()_=+~§±#{}-]`), "at least 1 special character"},
}

// PasswordProblems describes what a new password is missing ("" when it is
// acceptable). Rules match legacy minRequirementsMet, including its
// "longer than 8 characters" length rule.
func PasswordProblems(pw string) string {
	var missing []string
	for _, r := range passwordRules {
		if !r.re.MatchString(pw) {
			missing = append(missing, r.msg)
		}
	}
	if len(pw) <= 8 {
		missing = append(missing, "more than 8 characters")
	}
	if len(missing) == 0 {
		return ""
	}
	return "password needs " + strings.Join(missing, ", ")
}

// Login checks credentials. A reset-flagged account (Task 21: only after a
// correct password) gets a one-hour reset link instead of a session.
func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	ctx, span := tracer.Start(ctx, "auth.Service.Login")
	defer span.End()
	p := s.passwords
	u, reset, err := s.users.VerifyUser(ctx, user.User{Email: strings.TrimSpace(email), Password: null.StringFrom(password)},
		p.Iterations, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength)
	if err != nil && !reset {
		slog.InfoContext(ctx, fmt.Sprintf("failed login for %q: %v", email, err))
		return LoginResult{}, ErrInvalidCredentials
	}
	if reset {
		token := uuid.NewString()
		if err = s.tokens.Set(ctx, token, u.ID, loginResetTTL); err != nil {
			return LoginResult{}, err
		}
		return LoginResult{ResetRequired: true, ResetURL: "/reset/" + token}, nil
	}
	return LoginResult{User: u}, nil
}

func checkNew(newPassword, confirm string) error {
	f := svcerr.Fields{}
	if newPassword != confirm {
		f.Add("confirmationPassword", "passwords do not match")
	}
	if msg := PasswordProblems(newPassword); msg != "" {
		f.Add("newPassword", msg)
	}
	return f.Err()
}

func (s *Service) setPassword(ctx context.Context, u user.User, newPassword string) error {
	p := s.passwords
	u.Password = null.StringFrom(newPassword)
	if err := s.users.EditUserPassword(ctx, u, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength); err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}
	return nil
}

// ChangePassword changes the logged-in user's password after re-checking the old one.
func (s *Service) ChangePassword(ctx context.Context, u user.User, oldPassword, newPassword, confirm string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.ChangePassword")
	defer span.End()
	p := s.passwords
	check := u
	check.Password = null.StringFrom(oldPassword)
	if _, _, err := s.users.VerifyUser(ctx, check, p.Iterations, p.ScryptWorkFactor, p.ScryptBlockSize, p.ScryptParallelismFactor, p.KeyLength); err != nil {
		return svcerr.InvalidField("oldPassword", "old password is not correct")
	}
	if err := checkNew(newPassword, confirm); err != nil {
		return err
	}
	return s.setPassword(ctx, u, newPassword)
}

func (s *Service) tokenUser(ctx context.Context, token string) (user.User, error) {
	id, ok := s.tokens.Get(ctx, token)
	if !ok {
		return user.User{}, svcerr.NotFound("reset link is invalid or has expired", nil)
	}
	u, err := s.users.GetUser(ctx, user.User{ID: id})
	if err != nil || u.ID != id {
		s.tokens.Delete(ctx, token)
		return user.User{}, svcerr.NotFound("reset link is invalid or has expired", err)
	}
	return u, nil
}

// CheckResetToken reports NotFound for an unknown or expired token.
func (s *Service) CheckResetToken(ctx context.Context, token string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.CheckResetToken")
	defer span.End()
	_, err := s.tokenUser(ctx, token)
	return err
}

// ResetPassword sets a new password using a single-use reset token.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword, confirm string) error {
	ctx, span := tracer.Start(ctx, "auth.Service.ResetPassword")
	defer span.End()
	u, err := s.tokenUser(ctx, token)
	if err != nil {
		return err
	}
	if err = checkNew(newPassword, confirm); err != nil {
		return err
	}
	if err = s.setPassword(ctx, u, newPassword); err != nil {
		return err
	}
	s.tokens.Delete(ctx, token)
	return nil
}
