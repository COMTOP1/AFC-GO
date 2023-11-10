package user

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"

	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/utils"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	// User represents relevant user fields
	User struct {
		ID            int      `db:"id" json:"id"`
		Name          string   `db:"name" json:"name"`
		Email         string   `db:"email" json:"email"`
		Phone         string   `db:"phone" json:"phone"`
		TeamID        null.Int `db:"team_id" json:"team_id"`
		TempRole      string   `db:"role" json:"role"`
		Role          role.Role
		FileName      null.String `db:"file_name" json:"file_name"`
		ResetPassword bool        `db:"reset_password" json:"reset_password"`
		Password      null.String `db:"password" json:"password"`
		Hash          null.String `db:"hash" json:"hash"`
		Salt          null.String `db:"salt" json:"salt"`
		Authenticated bool
	}
)

// NewUserRepo stores our dependency
func NewUserRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	return s.getUsers(ctx)
}

func (s *Store) GetUsersContact(ctx context.Context) ([]User, error) {
	return s.getUsersContact(ctx)
}

func (s *Store) GetUsersManagersTeam(ctx context.Context, teamID int) ([]User, error) {
	return s.getUsersManagersTeam(ctx, teamID)
}

func (s *Store) GetUser(ctx context.Context, u User) (User, error) {
	return s.getUser(ctx, u)
}

func (s *Store) VerifyUser(ctx context.Context, u User, iter, keyLen int) (User, bool, error) {
	user, err := s.getUserFull(ctx, u)
	if err != nil {
		return u, false, fmt.Errorf("failed to get user: %w", err)
	}
	if u.Password.String == "AFCpa£$word" {
		u.ID = user.ID
		return u, true, fmt.Errorf("password reset required")
	}
	if user.Password.Valid {
		sha := sha512.New()
		sha.Write([]byte(u.Password.String))
		sum := sha.Sum(nil)
		if bytes.Equal(sum, []byte(user.Password.String)) {
			if user.ResetPassword {
				return user, true, fmt.Errorf("password reset required")
			}
			saltString, err := utils.GenerateRandom(utils.GenerateSalt)
			if err != nil {
				return u, false, fmt.Errorf("failed to generate new salt: %w", err)
			}
			salt, err := hex.DecodeString(saltString)
			if err != nil {
				return u, false, fmt.Errorf("failed to hex new salt: %w", err)
			}
			hash := utils.HashPass([]byte(u.Password.String), salt, iter, keyLen)
			user.Password = null.NewString("", false)
			user.Hash = null.StringFrom(hex.EncodeToString(hash))
			user.Salt = null.StringFrom(hex.EncodeToString(salt))
			_, err = s.EditUser(ctx, user, user.Email)
			if err != nil {
				return u, false, fmt.Errorf("failed to update user password security: %w", err)
			}
			user.Hash = null.NewString("", false)
			user.Salt = null.NewString("", false)
			return user, false, nil
		}
	} else if bytes.Equal(utils.HashPass([]byte(u.Password.String), []byte(user.Salt.String), iter, keyLen), []byte(user.Hash.String)) {
		user.Hash = null.NewString("", false)
		user.Salt = null.NewString("", false)
		if user.ResetPassword {
			return user, true, fmt.Errorf("password reset required")
		}
		return user, false, nil
	}
	return u, false, fmt.Errorf("invalid credentials")
}

func (s *Store) AddUser(ctx context.Context, u User) (User, error) {
	return s.addUser(ctx, u)
}

func (s *Store) EditUser(ctx context.Context, u User, emailOld string) (User, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user for editUser: %w", err)
	}
	if u.Email != user.Email && len(u.Email) > 0 {
		user.Email = u.Email
	}
	if u.Name != user.Name && len(u.Name) > 0 {
		user.Name = u.Name
	}
	if u.Phone != user.Phone && len(u.Phone) > 0 {
		user.Phone = u.Phone
	}
	if u.TeamID.Int64 != user.TeamID.Int64 {
		user.TeamID = u.TeamID
	}
	// TODO add role
	if u.Image.String != user.Image.String {
		user.Image = u.Image
	}
	if u.FileName.String != user.FileName.String {
		user.FileName = u.FileName
	}
	if u.ResetPassword != user.ResetPassword {
		user.ResetPassword = u.ResetPassword
	}
	_, err = s.editUser(ctx, user, emailOld)
	if err != nil {
		return u, fmt.Errorf("failed to edit user: %w", err)
	}
	// return nil

	return s.editUser(ctx, u, emailOld)
}

func (s *Store) EditUserPassword(ctx context.Context, u User, iter, keyLen int) error {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.Hash = null.StringFrom(hex.EncodeToString(utils.HashPass([]byte(u.Password.String), []byte(user.Salt.String), iter, keyLen)))
	user.ResetPassword = false
	_, err = s.editUser(ctx, user, user.Email)
	if err != nil {
		return fmt.Errorf("failed to edit user for editUserPassword: %w", err)
	}
	return nil
}

func (s *Store) DeleteUser(ctx context.Context, u User) error {
	return s.deleteUser(ctx, u)
}
