package utils

import sq "github.com/Masterminds/squirrel"

func MySQL() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Question)
}
