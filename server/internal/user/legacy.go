package user

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

// ClearLegacyPasswords removes the old site's unsalted SHA-512 password
// digests, along with any hash and salt on those accounts, flags each
// account for reset and signs it out everywhere. Those users can no longer sign in until they follow a
// reset link. It returns the IDs it cleared; running it again clears nothing.
func (s *Store) ClearLegacyPasswords(ctx context.Context) ([]int, error) {
	ctx, span := tracer.Start(ctx, "user.ClearLegacyPasswords")
	defer span.End()

	query, args, err := utils.PSQL().Update("users").
		Set("password", nil).
		Set("hash", nil).
		Set("salt", nil).
		Set("reset_password", true).
		Set("session_version", sq.Expr("session_version + 1")).
		Where(sq.NotEq{"password": nil}).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to build sql for clearing legacy passwords: %w", err)
	}
	var ids []int
	if err = s.db.SelectContext(ctx, &ids, query, args...); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to clear legacy passwords: %w", err)
	}
	return ids, nil
}
