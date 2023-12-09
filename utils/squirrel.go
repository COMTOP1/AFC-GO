package utils

import (
	sq1 "github.com/COMTOP1/AFC-GO/utils/squirrel"
	sq "github.com/Masterminds/squirrel"
)

func PSQL() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func PSQL1() sq1.StatementBuilderType {
	return sq1.StatementBuilder.PlaceholderFormat(sq1.Dollar)
}
