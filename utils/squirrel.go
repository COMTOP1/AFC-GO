package utils

import sq "github.com/Masterminds/squirrel"

func PSQL() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
