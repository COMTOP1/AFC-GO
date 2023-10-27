package user

import (
	"context"
	"fmt"
	"github.com/COMTOP1/AFC-GO/utils"
	sq "github.com/Masterminds/squirrel"
	"strconv"
)

func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	var u []User
	builder := sq.Select("id", "name", "email", "phone", "team_id", "role" /*"image",*/, "file_name").
		From("afc.users").
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsers: %w", err))
	}
	err = s.db.SelectContext(ctx, &u, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return u, nil
}

func (s *Store) getUsersContact(ctx context.Context) ([]User, error) {
	var u []User
	builder := sq.Select("id", "name", "email", "role").
		From("afc.users").
		Where("role IN ('PROGRAMME_EDITOR', 'LEAGUE_SECRETARY', 'TREASURER', 'SAFEGUARDING_OFFICER', 'CLUB_SECRETARY', 'CHAIRPERSON')").
		OrderBy("FIELD(role, 'PROGRAMME_EDITOR', 'LEAGUE_SECRETARY', 'TREASURER', 'SAFEGUARDING_OFFICER', 'CLUB_SECRETARY', 'CHAIRPERSON') DESC")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsersContact: %w", err))
	}
	err = s.db.SelectContext(ctx, &u, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return u, nil
}

func (s *Store) getUsersManagersTeam(ctx context.Context, teamID int) ([]User, error) {
	var p []User
	builder := sq.Select("id", "name", "image", "file_name").
	var u []User
		From("afc.users").
		Where(sq.Eq{"team_id": strconv.FormatUint(uint64(teamID), 10)}).
		OrderBy("id")
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsersManagersTeam: %w", err))
	}
	err = s.db.SelectContext(ctx, &u, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return u, nil
}

func (s *Store) getUser(ctx context.Context, u User) (User, error) {
	var u1 User
	builder := sq.Select("id", "name", "email", "phone", "team_id", "role" /*"image",*/, "file_name").
		From("afc.users").
		Where(sq.And{sq.Eq{"email": u.Email}, sq.NotEq{"email": ""}},
			sq.Eq{"id": u.ID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUser: %w", err))
	}
	err = s.db.GetContext(ctx, &u1, sql, args...)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return p1, nil
}

func (s *Store) addUser(ctx context.Context, u User) (User, error) {
	builder := utils.MySQL().Insert("afc.users").
		Columns("name", "email", "phone", "team_id", "role" /*"image",*/, "file_name").
		Values(u.Name, u.Email, u.Phone, u.TeamID, u.Role.DBString(), u.Image, u.FileName)
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
	u.ID = int(id)
	return u, nil
}

func (s *Store) editUser(ctx context.Context, u User) (User, error) {
	builder := utils.MySQL().Update("afc.users").
		SetMap(map[string]interface{}{
			"name":      u.Name,
			"email":     u.Email,
			"phone":     u.Phone,
			"team_id":   u.TeamID,
			"role":      u.Role.DBString(),
			"image":     u.Image,
			"file_name": u.FileName,
			"password":  u.Password,
			"hash":      u.Hash,
			"salt":      u.Salt,
		}).
		Where(sq.Eq{"email": emailOld})
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
	return u, nil
}

func (s *Store) deleteUser(ctx context.Context, u User) error {
	builder := utils.MySQL().Delete("afc.users").
		Where(sq.Eq{"email": u.Email})
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
