package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/server/internal/emails"
	"github.com/COMTOP1/AFC-GO/server/internal/role"
	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
	"github.com/COMTOP1/AFC-GO/server/internal/team"
	"github.com/COMTOP1/AFC-GO/server/internal/upload"
	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

const (
	uploadCategory = "user"
	resetTokenTTL  = 7 * 24 * time.Hour
)

type store interface {
	GetUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, userParam User) (User, error)
	AddUser(ctx context.Context, userParam User) (User, error)
	EditUser(ctx context.Context, userParam User) (User, error)
	DeleteUser(ctx context.Context, userParam User) error
}

// TeamGetter checks manager teams (satisfied by *team.Store).
type TeamGetter interface {
	GetTeam(ctx context.Context, teamParam team.Team) (team.Team, error)
}

// TokenSetter issues password reset tokens (satisfied by *auth.Tokens).
type TokenSetter interface {
	Set(ctx context.Context, token string, userID int, ttl time.Duration) error
}

// HashParams are the scrypt parameters for new users' initial passwords.
type HashParams struct {
	WorkFactor  int
	BlockSize   int
	Parallelism int
	KeyLength   int
}

// Service is user administration, shared by the API and legacy views.
type Service struct {
	store  store
	teams  TeamGetter
	files  *upload.Files
	mailer emails.Sender
	tokens TokenSetter
	hash   HashParams
	domain string
}

func NewService(store store, teams TeamGetter, files *upload.Files, mailer emails.Sender, tokens TokenSetter,
	hash HashParams, domain string) *Service {
	return &Service{store: store, teams: teams, files: files, mailer: mailer, tokens: tokens, hash: hash, domain: domain}
}

func (s *Service) admin(u User) Admin {
	return Admin{
		ID:       u.ID,
		Name:     u.Name,
		Email:    u.Email,
		Phone:    u.Phone.String,
		Role:     u.Role.String(),
		RoleCode: strings.ToLower(u.Role.DBString()),
		TeamID:   u.TeamID,
		ImageURL: s.files.URL(u.FileName.String),
	}
}

func (s *Service) get(ctx context.Context, id int) (User, error) {
	u, err := s.store.GetUser(ctx, User{ID: id})
	if err == nil && u.ID != id {
		// GetUser matches email OR id; a blank-email row must not stand in.
		return User{}, svcerr.NotFound("user not found", nil)
	}
	if err != nil {
		return User{}, svcerr.FromStore(err, "user")
	}
	return u, nil
}

func (s *Service) List(ctx context.Context) ([]Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.List")
	defer span.End()
	rows, err := s.store.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	out := make([]Admin, 0, len(rows))
	for _, u := range rows {
		out = append(out, s.admin(u))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, id int) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Get")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	return s.admin(u), nil
}

// validate checks u and normalises its team: only managers keep one.
func (s *Service) validate(ctx context.Context, u *User, roleCode string) error {
	f := svcerr.Fields{}
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	if u.Name == "" {
		f.Add("name", "name is required")
	}
	if !emailverifier.IsAddressValid(u.Email) {
		f.Add("email", "email address is not valid")
	}
	if roleCode != "" {
		r, err := role.GetRole(roleCode)
		if err != nil {
			f.Add("role", "unknown role")
		} else {
			u.Role = r
		}
	}
	if u.Role == role.Manager {
		if _, err := s.teams.GetTeam(ctx, team.Team{ID: u.TeamID}); err != nil {
			f.Add("teamId", "managers need an existing team")
		}
	} else {
		u.TeamID = 0
	}
	return f.Err()
}

func conflictOnDuplicate(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return svcerr.Conflict("email address is already in use", err)
	}
	return err
}

// Create adds a user with a generated password they must reset on first
// login, and emails it to them. If the email can't be sent the password is
// returned so the admin can pass it on.
func (s *Service) Create(ctx context.Context, in CreateInput, image *upload.File) (Created, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Create")
	defer span.End()

	phone := strings.TrimSpace(in.Phone)
	u := User{Name: in.Name, Email: in.Email, Phone: null.NewString(phone, phone != ""), TeamID: in.TeamID, ResetPassword: true}
	if in.Role == "" {
		return Created{}, svcerr.InvalidField("role", "role is required")
	}
	if err := s.validate(ctx, &u, in.Role); err != nil {
		return Created{}, err
	}

	password, err := utils.GenerateRandom(utils.GeneratePassword)
	if err != nil {
		return Created{}, fmt.Errorf("failed to generate password: %w", err)
	}
	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	if err != nil {
		return Created{}, fmt.Errorf("failed to generate salt: %w", err)
	}
	// Same hashing as legacy UserAddFunc (raw salt string); the user must
	// reset before this hash is ever checked.
	hash, err := utils.HashPassScrypt([]byte(password), []byte(salt), s.hash.WorkFactor, s.hash.BlockSize, s.hash.Parallelism, s.hash.KeyLength)
	if err != nil {
		return Created{}, fmt.Errorf("failed to hash password: %w", err)
	}
	u.Hash = null.StringFrom(hash)
	u.Salt = null.NewString(salt, salt != "")

	if image != nil {
		key, saveErr := s.files.Save(ctx, image, uploadCategory)
		if saveErr != nil {
			return Created{}, saveErr
		}
		u.FileName = null.StringFrom(key)
	}
	added, err := s.store.AddUser(ctx, u)
	if err != nil {
		s.files.Remove(ctx, u.FileName.String)
		return Created{}, conflictOnDuplicate(fmt.Errorf("failed to add user: %w", err))
	}

	out := Created{User: s.admin(added)}
	msg, err := emails.Signup(added.Email, added.Name, password, s.domain)
	if err == nil {
		err = s.mailer.Send(ctx, msg)
	}
	if err != nil {
		slog.InfoContext(ctx, fmt.Sprintf("signup email not sent to user %d: %+v", added.ID, err))
		out.TempPassword = password
		return out, nil
	}
	out.EmailSent = true
	return out, nil
}

func (s *Service) Update(ctx context.Context, id int, in UpdateInput, image *upload.File) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Update")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	if in.Name != nil {
		u.Name = *in.Name
	}
	if in.Email != nil {
		u.Email = *in.Email
	}
	if in.Phone != nil {
		phone := strings.TrimSpace(*in.Phone)
		u.Phone = null.NewString(phone, phone != "")
	}
	if in.TeamID != nil {
		u.TeamID = *in.TeamID
	}
	roleCode := ""
	if in.Role != nil {
		// A present-but-blank role is a caller mistake, not "leave it
		// alone" - that's expressed by leaving the field out entirely
		// (in.Role == nil).
		if *in.Role == "" {
			return Admin{}, svcerr.InvalidField("role", "role is required")
		}
		roleCode = *in.Role
	}
	if err = s.validate(ctx, &u, roleCode); err != nil {
		return Admin{}, err
	}

	oldKey, newKey := u.FileName.String, ""
	switch {
	case image != nil:
		if newKey, err = s.files.Save(ctx, image, uploadCategory); err != nil {
			return Admin{}, err
		}
		u.FileName = null.StringFrom(newKey)
	case in.RemoveImage:
		u.FileName = null.String{}
	}
	if _, err = s.store.EditUser(ctx, u); err != nil {
		s.files.Remove(ctx, newKey)
		return Admin{}, conflictOnDuplicate(fmt.Errorf("failed to edit user: %w", err))
	}
	if oldKey != "" && oldKey != u.FileName.String {
		s.files.Remove(ctx, oldKey)
	}
	return s.admin(u), nil
}

// Delete removes the user and their photo, returning what was deleted.
func (s *Service) Delete(ctx context.Context, id int) (Admin, error) {
	ctx, span := tracer.Start(ctx, "user.Service.Delete")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return Admin{}, err
	}
	deleted := s.admin(u)
	if err = s.store.DeleteUser(ctx, u); err != nil {
		return Admin{}, fmt.Errorf("failed to delete user: %w", err)
	}
	s.files.Remove(ctx, u.FileName.String)
	return deleted, nil
}

// ResetPassword forces a reset on next login and emails a 7-day reset link.
// If the email can't be sent, the link is returned for the admin to pass on.
func (s *Service) ResetPassword(ctx context.Context, id int) (ResetResult, error) {
	ctx, span := tracer.Start(ctx, "user.Service.ResetPassword")
	defer span.End()
	u, err := s.get(ctx, id)
	if err != nil {
		return ResetResult{}, err
	}
	u.ResetPassword = true
	if _, err = s.store.EditUser(ctx, u); err != nil {
		return ResetResult{}, fmt.Errorf("failed to flag user for reset: %w", err)
	}
	token := uuid.NewString()
	if err = s.tokens.Set(ctx, token, u.ID, resetTokenTTL); err != nil {
		return ResetResult{}, fmt.Errorf("failed to store reset token: %w", err)
	}
	link := fmt.Sprintf("https://%s/reset/%s", s.domain, token)

	msg, err := emails.Reset(u.Email, link)
	if err == nil {
		err = s.mailer.Send(ctx, msg)
	}
	if err != nil {
		slog.InfoContext(ctx, fmt.Sprintf("reset email not sent to user %d: %+v", u.ID, err))
		return ResetResult{ResetURL: link}, nil
	}
	return ResetResult{EmailSent: true}, nil
}
