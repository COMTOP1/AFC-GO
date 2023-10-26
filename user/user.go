package user

import (
	"context"
	"github.com/COMTOP1/AFC-GO/role"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	User struct {
		ID            int      `db:"id" json:"id"`
		Name          string   `db:"name" json:"name"`
		Email         string   `db:"email" json:"email"`
		Phone         string   `db:"phone" json:"phone"`
		TeamID        null.Int `db:"team_id" json:"team_id"`
		TempRole      string   `db:"role" json:"role"`
		Role          role.Role
		Image         null.String `db:"image" json:"image"`
		FileName      null.String `db:"file_name" json:"file_name"`
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

func (s *Store) AddUser(ctx context.Context, p User) (User, error) {
	return s.addUser(ctx, p)
func (s *Store) AddUser(ctx context.Context, u User) (User, error) {
	return s.addUser(ctx, u)
}

func (s *Store) EditUser(ctx context.Context, u User, emailOld string) (User, error) {
	return s.editUser(ctx, u, emailOld)
}

func (s *Store) DeleteUser(ctx context.Context, u User) error {
	return s.deleteUser(ctx, u)
}
