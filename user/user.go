package user

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
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
		ID            int         `db:"id" json:"id"`
		Name          string      `db:"name" json:"name"`
		Email         string      `db:"email" json:"email"`
		Phone         null.String `db:"phone" json:"phone"`
		TeamID        int         `db:"team_id" json:"team_id"`
		TempRole      string      `db:"role" json:"role"`
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

func (s *Store) VerifyUser(ctx context.Context, userParam User, iter, workFactor, blockSize, parallelismFactor, keyLen int) (User, bool, error) {
	user, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return userParam, false, fmt.Errorf("failed to get user: %w", err)
	}
	if user.ResetPassword == true {
		userParam.ID = user.ID
		return userParam, true, errors.New("password reset required")
	}
	var hashDecode, saltDecode []byte
	if user.Hash.Valid {
		hashDecode, err = hex.DecodeString(user.Hash.String)
		if err != nil {
			return userParam, false, fmt.Errorf("failed to decode hex of hash verify user: %w", err)
		}
	}
	if user.Salt.Valid {
		saltDecode, err = hex.DecodeString(user.Salt.String)
		if err != nil {
			return userParam, false, fmt.Errorf("failed to decode hex of salt verify user: %w", err)
		}
	}
	if user.Password.Valid {
		sha := sha512.New()
		sha.Write([]byte(userParam.Password.String))
		sum := sha.Sum(nil)
		if bytes.Equal(sum, []byte(user.Password.String)) {
			if user.ResetPassword {
				return user, true, errors.New("password reset required")
			}
			var salt string
			salt, err = utils.GenerateRandom(utils.GenerateSalt)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to generate new salt: %w", err)
			}
			var hash string
			hash, err = utils.HashPassScrypt([]byte(userParam.Password.String), []byte(salt), workFactor, blockSize, parallelismFactor, keyLen)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to generate password hash: %w", err)
			}
			user.Password = null.NewString("", false)
			user.Hash = null.StringFrom(hash)
			user.Salt = null.StringFrom(hex.EncodeToString([]byte(salt)))
			err = s.editUser(ctx, user)
			if err != nil {
				return userParam, false, fmt.Errorf("failed to update user password security: %w", err)
			}
			user.Hash = null.NewString("", false)
			user.Salt = null.NewString("", false)
			return user, false, nil
		}
		return userParam, false, errors.New("invalid credentials")
	} else if bytes.Equal(utils.HashPass([]byte(userParam.Password.String), saltDecode, iter, keyLen), hashDecode) {
		var hash string
		hash, err = utils.HashPassScrypt([]byte(userParam.Password.String), saltDecode, workFactor, blockSize, parallelismFactor, keyLen)
		if err != nil {
			return userParam, false, fmt.Errorf("failed to generate password hash: %w", err)
		}
		user.Password = null.NewString("", false)
		user.Hash = null.StringFrom(hash)
		user.Salt = null.StringFrom(hex.EncodeToString(saltDecode))
		err = s.editUser(ctx, user)
		if err != nil {
			return userParam, false, fmt.Errorf("failed to update user password security: %w", err)
		}
		user.Hash = null.NewString("", false)
		user.Salt = null.NewString("", false)
		if user.ResetPassword {
			return user, true, errors.New("password reset required")
		}
		return user, false, nil
	}
	scryptHash, err := utils.HashPassScrypt([]byte(userParam.Password.String), saltDecode, workFactor, blockSize, parallelismFactor, keyLen)
	if err != nil {
		return userParam, false, fmt.Errorf("failed to generate password hash verify: %w", err)
	}
	if scryptHash == user.Hash.String {
		user.Hash = null.NewString("", false)
		user.Salt = null.NewString("", false)
		if user.ResetPassword {
			return user, true, errors.New("password reset required")
		}
		return user, false, nil
	}
	return userParam, false, errors.New("invalid credentials")
}

func (s *Store) AddUser(ctx context.Context, userParam User) (User, error) {
	return s.addUser(ctx, userParam)
}

func (s *Store) EditUser(ctx context.Context, userParam User) (User, error) {
	userDB, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return userParam, fmt.Errorf("failed to get user for edit user: %w", err)
	}
	if userParam.Email != userDB.Email && len(userParam.Email) > 0 {
		userDB.Email = userParam.Email
	}
	if userParam.Name != userDB.Name && len(userParam.Name) > 0 {
		userDB.Name = userParam.Name
	}
	if userParam.Phone.Valid && (!userDB.Phone.Valid || userDB.Phone.String != userParam.Phone.String) {
		userDB.Phone = userParam.Phone
	}
	if userParam.TeamID != userDB.TeamID {
		userDB.TeamID = userParam.TeamID
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
	err = s.editUser(ctx, userDB)
	if err != nil {
		return userParam, fmt.Errorf("failed to edit user: %w", err)
	}
	return userParam, nil
}

func (s *Store) EditUserPassword(ctx context.Context, userParam User, workFactor, blockSize, parallelismFactor, keyLen int) error {
	user, err := s.getUserFull(ctx, userParam)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	var saltDecode []byte
	if user.Salt.Valid {
		saltDecode, err = hex.DecodeString(user.Salt.String)
		if err != nil {
			return fmt.Errorf("failed to decode hex of salt verify user: %w", err)
		}
	}

	scryptHash, err := utils.HashPassScrypt([]byte(userParam.Password.String), saltDecode, workFactor, blockSize, parallelismFactor, keyLen)
	if err != nil {
		return fmt.Errorf("failed to generate password hash verify: %w", err)
	}

	user.Hash = null.StringFrom(scryptHash)
	user.ResetPassword = false
	user.Role = userParam.Role

	err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for edit user password: %w", err)
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
	err = s.editUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to edit user for edit user password: %w", err)
	}
	return nil
}

func (s *Store) DeleteUser(ctx context.Context, userParam User) error {
	return s.deleteUser(ctx, userParam)
}
