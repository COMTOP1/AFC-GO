package user

import (
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/COMTOP1/AFC-GO/server/internal/utils"
)

// wrappedLegacyPrefix marks a password column holding scrypt(legacy digest)
// rather than the bare, unsalted SHA-512 digest the old site stored. The
// salt for the scrypt layer is in the salt column.
const wrappedLegacyPrefix = "scrypt-sha512:"

// legacyMatches checks password against a legacy password column: either
// the original unsalted SHA-512 digest, or that digest wrapped in scrypt by
// WrapLegacyPasswords.
func legacyMatches(password, stored string, salt []byte, workFactor, blockSize, parallelismFactor, keyLen int) (bool, error) {
	digest := sha512.Sum512([]byte(password))
	wrapped, ok := strings.CutPrefix(stored, wrappedLegacyPrefix)
	if !ok {
		return subtle.ConstantTimeCompare(digest[:], []byte(stored)) == 1, nil
	}
	hash, err := utils.HashPassScrypt(digest[:], salt, workFactor, blockSize, parallelismFactor, keyLen)
	if err != nil {
		return false, fmt.Errorf("failed to hash legacy password: %w", err)
	}
	return subtle.ConstantTimeCompare([]byte(hash), []byte(wrapped)) == 1, nil
}

// WrapLegacyPasswords replaces every bare legacy SHA-512 digest with
// scrypt(digest), so the database no longer holds fast, unsalted hashes for
// accounts that haven't logged in since the old site. Users keep their
// passwords; VerifyUser upgrades them to a plain scrypt hash on their next
// login. It is safe to run more than once and returns how many rows it
// wrapped.
func (s *Store) WrapLegacyPasswords(ctx context.Context, workFactor, blockSize, parallelismFactor, keyLen int) (int, error) {
	ctx, span := tracer.Start(ctx, "user.WrapLegacyPasswords")
	defer span.End()

	query, args, err := utils.PSQL().Select("id", "password").
		From("users").
		Where(sq.And{sq.NotEq{"password": nil}, sq.NotLike{"password": wrappedLegacyPrefix + "%"}}).
		ToSql()
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to build sql for legacy passwords: %w", err)
	}
	var rows []struct {
		ID       int    `db:"id"`
		Password string `db:"password"`
	}
	if err = s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get legacy passwords: %w", err)
	}

	wrapped := 0
	for _, row := range rows {
		var salt, hash string
		salt, err = utils.GenerateRandom(utils.GenerateSalt)
		if err != nil {
			return wrapped, fmt.Errorf("failed to generate salt: %w", err)
		}
		hash, err = utils.HashPassScrypt([]byte(row.Password), []byte(salt), workFactor, blockSize, parallelismFactor, keyLen)
		if err != nil {
			return wrapped, fmt.Errorf("failed to wrap legacy password for user %d: %w", row.ID, err)
		}
		query, args, err = utils.PSQL().Update("users").
			Set("password", wrappedLegacyPrefix+hash).
			Set("salt", hex.EncodeToString([]byte(salt))).
			// Skip the row if the user logged in (and was upgraded) meanwhile.
			Where(sq.Eq{"id": row.ID, "password": row.Password}).
			ToSql()
		if err != nil {
			span.RecordError(err)
			return wrapped, fmt.Errorf("failed to build sql for wrapping legacy password: %w", err)
		}
		res, execErr := s.db.ExecContext(ctx, query, args...)
		if execErr != nil {
			span.RecordError(execErr)
			return wrapped, fmt.Errorf("failed to wrap legacy password for user %d: %w", row.ID, execErr)
		}
		if n, _ := res.RowsAffected(); n == 1 {
			wrapped++
		}
	}
	return wrapped, nil
}
