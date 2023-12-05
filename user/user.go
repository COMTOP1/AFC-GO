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
	"github.com/COMTOP1/AFC-GO/team"
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

func (s *Store) GetUsersManagersTeam(ctx context.Context, teamParam team.Team) ([]User, error) {
	return s.getUsersManagersTeam(ctx, teamParam)
}

func (s *Store) GetUser(ctx context.Context, userParam User) (User, error) {
	return s.getUser(ctx, userParam)
}

func (s *Store) VerifyUser(ctx context.Context, userParam User, iter, keyLen int) (User, bool, error) {
	user, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return userParam, false, fmt.Errorf("failed to get user: %w", err)
	}
	if userParam.Password.String == "AFCpa£$word" {
		userParam.ID = user.ID
		return userParam, true, fmt.Errorf("password reset required")
	}
	if user.Password.Valid {
		sha := sha512.New()
		sha.Write([]byte(userParam.Password.String))
		sum := sha.Sum(nil)
		if bytes.Equal(sum, []byte(user.Password.String)) {
			if user.ResetPassword {
				return user, true, fmt.Errorf("password reset required")
			}
			var saltString string
			saltString, err = utils.GenerateRandom(utils.GenerateSalt)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to generate new salt: %w", err)
			}
			var salt []byte
			salt, err = hex.DecodeString(saltString)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to hex new salt: %w", err)
			}
			hash := utils.HashPass([]byte(userParam.Password.String), salt, iter, keyLen)
			user.Password = null.NewString("", false)
			user.Hash = null.StringFrom(hex.EncodeToString(hash))
			user.Salt = null.StringFrom(hex.EncodeToString(salt))
			_, err = s.EditUser(ctx, user)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to update user password security: %w", err)
			}
			user.Hash = null.NewString("", false)
			user.Salt = null.NewString("", false)
			return user, false, nil
		}
	} else if bytes.Equal(utils.HashPass([]byte(userParam.Password.String), []byte(user.Salt.String), iter, keyLen), []byte(user.Hash.String)) {
		user.Hash = null.NewString("", false)
		user.Salt = null.NewString("", false)
		if user.ResetPassword {
			return user, true, fmt.Errorf("password reset required")
		}
		return user, false, nil
	}
	return userParam, false, fmt.Errorf("invalid credentials")
}

func (s *Store) AddUser(ctx context.Context, userParam User) (User, error) {
	return s.addUser(ctx, userParam)
}

func (s *Store) EditUser(ctx context.Context, userParam User) (User, error) {
	userDB, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return userParam, fmt.Errorf("failed to get user for editUser: %w", err)
	}
	if userParam.Email != userDB.Email && len(userParam.Email) > 0 {
		userDB.Email = userParam.Email
	}
	if userParam.Name != userDB.Name && len(userParam.Name) > 0 {
		userDB.Name = userParam.Name
	}
	if userParam.Phone != userDB.Phone && len(userParam.Phone) > 0 {
		userDB.Phone = userParam.Phone
	}
	if userParam.TeamID.Int64 != userDB.TeamID.Int64 {
		userDB.TeamID = userParam.TeamID
	}
	userDB.Role, err = role.GetRole(userDB.TempRole)
	if err != nil {
		return userParam, fmt.Errorf("failed to parse role for editUser: %w", err)
	}
	if userParam.Role != userDB.Role {
		userDB.Role = userParam.Role
	}
	if userParam.FileName.String != userDB.FileName.String {
		userDB.FileName = userParam.FileName
	}
	if userParam.ResetPassword != userDB.ResetPassword {
		userDB.ResetPassword = userParam.ResetPassword
	}
	_, err = s.editUser(ctx, userDB)
	if err != nil {
		return userParam, fmt.Errorf("failed to edit user: %w", err)
	}

	return s.editUser(ctx, userParam, emailOld)
}

func (s *Store) EditUserPassword(ctx context.Context, userParam User, iter, keyLen int) error {
	user, err := s.GetUser(ctx, userParam)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.Hash = null.StringFrom(hex.EncodeToString(utils.HashPass([]byte(userParam.Password.String), []byte(user.Salt.String), iter, keyLen)))
	user.ResetPassword = false
	user.Role = userParam.Role
	_, err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for editUserPassword: %w", err)
	}
	return nil
}

func (s *Store) EditUserImage(ctx context.Context, userParam User) error {
	user, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.Role = userParam.Role
	user.FileName = userParam.FileName
	_, err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for editUserPassword: %w", err)
	}
	return nil
}

func (s *Store) DeleteUser(ctx context.Context, userParam User) error {
	return s.deleteUser(ctx, userParam)
}
