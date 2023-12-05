package user

import (
	"context"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	var usersDB []User
	builder := sq.Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password").
		From("afc.users").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsers: %w", err))
	}
	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return usersDB, nil
}

func (s *Store) getUsersContact(ctx context.Context) ([]User, error) {
	var usersDB []User
	builder := sq.Select("id", "name", "email", "role").
		From("afc.users").
		Where("role IN ('PROGRAMME_EDITOR', 'LEAGUE_SECRETARY', 'TREASURER', 'SAFEGUARDING_OFFICER', 'CLUB_SECRETARY', 'CHAIRPERSON')").
		OrderBy("FIELD(role, 'PROGRAMME_EDITOR', 'LEAGUE_SECRETARY', 'TREASURER', 'SAFEGUARDING_OFFICER', 'CLUB_SECRETARY', 'CHAIRPERSON') DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsersContact: %w", err))
	}
	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return usersDB, nil
}

func (s *Store) getUsersManagersTeam(ctx context.Context, teamParam team.Team) ([]User, error) {
	var usersDB []User
	builder := utils.MySQL().Select("id", "name").
		From("afc.users").
		Where(sq.Eq{"team_id": strconv.FormatUint(uint64(teamParam.ID), 10)}).
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsersManagersTeam: %w", err))
	}
	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return usersDB, nil
}

func (s *Store) getUser(ctx context.Context, userParam User) (User, error) {
	var userDB User
	builder := utils.MySQL().Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password").
		From("afc.users").
		Where(sq.Or{sq.Eq{"email": userParam.Email},
			sq.Eq{"id": userParam.ID}})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUser: %w", err))
	}
	err = s.db.GetContext(ctx, &userDB, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userDB, nil
}

func (s *Store) getUserFull(ctx context.Context, userParam User) (User, error) {
	var userDB User
	builder := utils.MySQL().Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password", "password", "hash", "salt").
		From("afc.users").
		Where(sq.Or{sq.Eq{"email": userParam.Email},
			sq.Eq{"id": userParam.ID}})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUser: %w", err))
	}
	err = s.db.GetContext(ctx, &userDB, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userDB, nil
}

func (s *Store) addUser(ctx context.Context, userParam User) (User, error) {
	builder := utils.MySQL().Insert("afc.users").
		Columns("name", "email", "phone", "team_id", "role", "file_name", "reset_password", "hash", "salt").
		Values(userParam.Name, userParam.Email, userParam.Phone, userParam.TeamID, userParam.Role.DBString(), userParam.FileName, userParam.ResetPassword, userParam.Hash, userParam.Salt)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addUser: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	if rows < 1 {
		return User{}, fmt.Errorf("failed to add user: invalid rows affected: %d", rows)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	userParam.ID = int(id)
	return userParam, nil
}

func (s *Store) editUser(ctx context.Context, userParam User, emailOld string) (User, error) {
	builder := utils.MySQL().Update("afc.users").
		SetMap(map[string]interface{}{
			"name":           userParam.Name,
			"email":          userParam.Email,
			"phone":          userParam.Phone,
			"team_id":        userParam.TeamID,
			"role":           userParam.Role.DBString(),
			"file_name":      userParam.FileName,
			"reset_password": userParam.ResetPassword,
			"password":       userParam.Password,
			"hash":           userParam.Hash,
			"salt":           userParam.Salt,
		}).
		Where(sq.Eq{"id": userParam.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editUser: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to edit user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("failed to edit user: %w", err)
	}
	if rows < 1 {
		return User{}, fmt.Errorf("failed to edit user: invalid rows affected: %d, this user may not exist: %s", rows, emailOld)
	}
	return userParam, nil
}

func (s *Store) deleteUser(ctx context.Context, userParam User) error {
	builder := utils.MySQL().Delete("afc.users").
		Where(sq.Eq{"email": userParam.Email})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteUser: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
