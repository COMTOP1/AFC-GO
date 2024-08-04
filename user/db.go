package user

import (
	"context"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/role"
	"github.com/COMTOP1/AFC-GO/team"
	"github.com/COMTOP1/AFC-GO/utils"
)

func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	//nolint:prealloc
	var usersDB, users []User
	builder := sq.Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password").
		From("afc.users").
		OrderBy("id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get users: %w", err))
	}

	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range usersDB {
		user.Role, err = role.GetRole(user.TempRole)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *Store) getUsersContact(ctx context.Context) ([]User, error) {
	//nolint:prealloc
	var usersDB, users []User

	caseBuilder := sq.Case().
		When("role = 'CHAIRPERSON'", "1").
		When("role = 'CLUB_SECRETARY'", "2").
		When("role = 'SAFEGUARDING_OFFICER'", "3").
		When("role = 'TREASURER'", "4").
		When("role = 'LEAGUE_SECRETARY'", "5").
		When("role = 'PROGRAMME_EDITOR'", "6")

	caseSQL, _, err := caseBuilder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build case sql for get users contact: %w", err))
	}

	builder := sq.Select("id", "name", "email", "role").
		From("afc.users").
		Where("role IN ('PROGRAMME_EDITOR', 'LEAGUE_SECRETARY', 'TREASURER', 'SAFEGUARDING_OFFICER', 'CLUB_SECRETARY', 'CHAIRPERSON')").
		OrderBy(caseSQL)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get users contact: %w", err))
	}

	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	for _, user := range usersDB {
		user.Role, err = role.GetRole(user.TempRole)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *Store) getUsersManagersTeam(ctx context.Context, teamParam team.Team) ([]User, error) {
	var usersDB []User

	builder := utils.PSQL().Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password").
		From("afc.users").
		Where(sq.Eq{"team_id": strconv.FormatUint(uint64(teamParam.ID), 10)}).
		OrderBy("id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get users managers team: %w", err))
	}

	err = s.db.SelectContext(ctx, &usersDB, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return usersDB, nil
}

func (s *Store) getUser(ctx context.Context, userParam User) (User, error) {
	var userDB User

	builder := utils.PSQL().Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password").
		From("afc.users").
		Where(sq.Or{sq.Eq{"email": userParam.Email},
			sq.Eq{"id": userParam.ID}})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get user: %w", err))
	}

	err = s.db.GetContext(ctx, &userDB, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	userDB.Role, err = role.GetRole(userDB.TempRole)
	if err != nil {
		return User{}, fmt.Errorf("failed to parse role: %w", err)
	}

	userDB.TempRole = ""

	return userDB, nil
}

func (s *Store) getUserFull(ctx context.Context, userParam User) (User, error) {
	var userDB User

	builder := utils.PSQL().Select("id", "name", "email", "phone", "team_id", "role", "file_name", "reset_password", "password", "hash", "salt").
		From("afc.users").
		Where(sq.Or{sq.Eq{"email": userParam.Email},
			sq.Eq{"id": userParam.ID}})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for get user: %w", err))
	}

	err = s.db.GetContext(ctx, &userDB, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}

	userDB.Role, err = role.GetRole(userDB.TempRole)
	if err != nil {
		return User{}, fmt.Errorf("failed to parse role: %w", err)
	}

	userDB.TempRole = ""

	return userDB, nil
}

func (s *Store) addUser(ctx context.Context, userParam User) (User, error) {
	builder := utils.PSQL().Insert("afc.users").
		Columns("name", "email", "phone", "team_id", "role", "file_name", "reset_password", "hash", "salt").
		Values(userParam.Name, userParam.Email, userParam.Phone, userParam.TeamID, userParam.Role.DBString(), userParam.FileName, userParam.ResetPassword, userParam.Hash, userParam.Salt)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for add user: %w", err))
	}

	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}

	return userParam, nil
}

func (s *Store) editUser(ctx context.Context, userParam User) error {
	builder := utils.PSQL().Update("afc.users").
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
		panic(fmt.Errorf("failed to build sql for edit user: %w", err))
	}

	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to edit user: %w", err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to edit user: %w", err)
	}

	return nil
}

func (s *Store) deleteUser(ctx context.Context, userParam User) error {
	builder := utils.PSQL().Delete("afc.users").
		Where(sq.Eq{"email": userParam.Email})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for delete user: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
