package squirrel

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

type selectData struct {
	PlaceholderFormat sq.PlaceholderFormat
	RunWith           sq.BaseRunner
	Prefixes          []sq.Sqlizer
	Options           []string
	Columns           []sq.Sqlizer
	From              sq.Sqlizer
	Joins             []sq.Sqlizer
	WhereParts        []sq.Sqlizer
	GroupBys          []string
	HavingParts       []sq.Sqlizer
	OrderByParts      []sq.Sqlizer
	Limit             string
	Offset            string
	Union             []sq.Sqlizer
	UnionAll          []sq.Sqlizer
	Suffixes          []sq.Sqlizer
}

func (d *selectData) Exec() (sql.Result, error) {
	if d.RunWith == nil {
		return nil, sq.RunnerNotSet
	}
	return sq.ExecWith(d.RunWith, d)
}

func (d *selectData) Query() (*sql.Rows, error) {
	if d.RunWith == nil {
		return nil, sq.RunnerNotSet
	}
	return sq.QueryWith(d.RunWith, d)
}

func (d *selectData) QueryRow() sq.RowScanner {
	if d.RunWith == nil {
		return &Row{err: sq.RunnerNotSet}
	}
	queryRower, ok := d.RunWith.(sq.QueryRower)
	if !ok {
		return &Row{err: sq.RunnerNotQueryRunner}
	}
	return sq.QueryRowWith(queryRower, d)
}

//nolint:revive
func (d *selectData) ToSql() (sqlStr string, args []interface{}, err error) {
	sqlStr, args, err = d.toSqlRaw(false)
	if err != nil {
		return
	}

	sqlStr, err = d.PlaceholderFormat.ReplacePlaceholders(sqlStr)
	return
}

//nolint:revive
func (d *selectData) toSqlRaw(subQuery bool) (sqlStr string, args []interface{}, err error) {
	if len(d.Columns) == 0 {
		err = fmt.Errorf("select statements must have at least one result column")
		return
	}

	sql1 := &bytes.Buffer{}

	if len(d.Prefixes) > 0 {
		args, err = appendToSql(d.Prefixes, sql1, " ", args)
		if err != nil {
			return
		}

		sql1.WriteString(" ")
	}

	sql1.WriteString("(SELECT ")

	if len(d.Options) > 0 {
		sql1.WriteString(strings.Join(d.Options, " "))
		sql1.WriteString(" ")
	}

	if len(d.Columns) > 0 {
		args, err = appendToSql(d.Columns, sql1, ", ", args)
		if err != nil {
			return
		}
	}

	if d.From != nil {
		sql1.WriteString(" FROM ")
		args, err = appendToSql([]sq.Sqlizer{d.From}, sql1, "", args)
		if err != nil {
			return
		}
	}

	if len(d.Joins) > 0 {
		sql1.WriteString(" ")
		args, err = appendToSql(d.Joins, sql1, " ", args)
		if err != nil {
			return
		}
	}

	if len(d.WhereParts) > 0 {
		sql1.WriteString(" WHERE ")
		args, err = appendToSql(d.WhereParts, sql1, " AND ", args)
		if err != nil {
			return
		}
	}

	if len(d.GroupBys) > 0 {
		sql1.WriteString(" GROUP BY ")
		sql1.WriteString(strings.Join(d.GroupBys, ", "))
	}

	if len(d.HavingParts) > 0 {
		sql1.WriteString(" HAVING ")
		args, err = appendToSql(d.HavingParts, sql1, " AND ", args)
		if err != nil {
			return
		}
	}

	if len(d.OrderByParts) > 0 {
		sql1.WriteString(" ORDER BY ")
		args, err = appendToSql(d.OrderByParts, sql1, ", ", args)
		if err != nil {
			return
		}
	}

	if len(d.Limit) > 0 {
		sql1.WriteString(" LIMIT ")
		sql1.WriteString(d.Limit)
	}

	if len(d.Offset) > 0 {
		sql1.WriteString(" OFFSET ")
		sql1.WriteString(d.Offset)
	}

	if len(d.Union) > 0 {
		sql1.WriteString(" UNION ")
		args, err = appendToSql(d.Union, sql1, " UNION ", args)
		if err != nil {
			return
		}
	}

	if len(d.UnionAll) > 0 {
		sql1.WriteString(") UNION ALL ")
		args, err = appendToSql(d.UnionAll, sql1, ") UNION ALL ", args)
		if err != nil {
			return
		}
	}

	if len(d.Suffixes) > 0 {
		sql1.WriteString(" ")

		args, err = appendToSql(d.Suffixes, sql1, " ", args)
		if err != nil {
			return
		}
	}

	if subQuery {
		sql1.WriteString(")")
	}

	sqlStr = sql1.String()
	return
}

// Builder

// SelectBuilder builds SQL SELECT statements.
type SelectBuilder builder.Builder

func init() {
	builder.Register(SelectBuilder{}, selectData{})
}

// Format methods

// PlaceholderFormat sets PlaceholderFormat (e.g. Question or Dollar) for the
// query.
func (b SelectBuilder) PlaceholderFormat(f PlaceholderFormat) SelectBuilder {
	return builder.Set(b, "PlaceholderFormat", f).(SelectBuilder)
}

// Runner methods

// RunWith sets a Runner (like database/sql.DB) to be used with e.g. Exec.
// For most cases runner will be a database connection.
//
// Internally, we use this to mock out the database connection for testing.
func (b SelectBuilder) RunWith(runner BaseRunner) SelectBuilder {
	return setRunWith(b, runner).(SelectBuilder)
}

// Exec builds and Execs the query with the Runner set by RunWith.
func (b SelectBuilder) Exec() (sql.Result, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Exec()
}

// Query builds and queries the query with the Runner set by RunWith.
func (b SelectBuilder) Query() (*sql.Rows, error) {
	data := builder.GetStruct(b).(selectData)
	return data.Query()
}

// QueryRow builds and QueryRows the query with the Runner set by RunWith.
func (b SelectBuilder) QueryRow() RowScanner {
	data := builder.GetStruct(b).(selectData)
	return data.QueryRow()
}

// Scan is a shortcut for QueryRow().Scan.
func (b SelectBuilder) Scan(dest ...interface{}) error {
	return b.QueryRow().Scan(dest...)
}

// SQL methods

// ToSql builds the query into a SQL string and bound args.
//
//nolint:revive
func (b SelectBuilder) ToSql() (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.ToSql()
}

//nolint:revive
func (b SelectBuilder) toSqlRaw() (string, []interface{}, error) {
	data := builder.GetStruct(b).(selectData)
	return data.toSqlRaw(false)
}

// MustSql builds the query into a SQL string and bound args.
// It panics if there are any errors.
//
//nolint:revive
func (b SelectBuilder) MustSql() (string, []interface{}) {
	sql1, args, err := b.ToSql()
	if err != nil {
		panic(err)
	}
	return sql1, args
}

// Prefix adds an expression to the beginning of the query
func (b SelectBuilder) Prefix(sql string, args ...interface{}) SelectBuilder {
	return b.PrefixExpr(sq.Expr(sql, args...))
}

// PrefixExpr adds an expression to the very beginning of the query
func (b SelectBuilder) PrefixExpr(expr Sqlizer) SelectBuilder {
	return builder.Append(b, "Prefixes", expr).(SelectBuilder)
}

// Distinct adds a DISTINCT clause to the query.
func (b SelectBuilder) Distinct() SelectBuilder {
	return b.Options("DISTINCT")
}

// Options adds select option to the query
func (b SelectBuilder) Options(options ...string) SelectBuilder {
	return builder.Extend(b, "Options", options).(SelectBuilder)
}

// Columns adds result columns to the query.
func (b SelectBuilder) Columns(columns ...string) SelectBuilder {
	parts := make([]interface{}, 0, len(columns))
	for _, str := range columns {
		parts = append(parts, newPart(str))
	}
	return builder.Extend(b, "Columns", parts).(SelectBuilder)
}

// RemoveColumns remove all columns from query.
// Must add a new column with Column or Columns methods, otherwise
// return a error.
func (b SelectBuilder) RemoveColumns() SelectBuilder {
	return builder.Delete(b, "Columns").(SelectBuilder)
}

// Column adds a result column to the query.
// Unlike Columns, Column accepts args which will be bound to placeholders in
// the columns string, for example:
//
//	Column("IF(col IN ("+squirrel.Placeholders(3)+"), 1, 0) as col", 1, 2, 3)
func (b SelectBuilder) Column(column interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Columns", newPart(column, args...)).(SelectBuilder)
}

// From sets the FROM clause of the query.
func (b SelectBuilder) From(from string) SelectBuilder {
	return builder.Set(b, "From", newPart(from)).(SelectBuilder)
}

// FromSelect sets a sub-query into the FROM clause of the query.
func (b SelectBuilder) FromSelect(from SelectBuilder, alias string) SelectBuilder {
	// Prevent misnumbered parameters in nested selects (#183).
	from = from.PlaceholderFormat(Dollar)
	return builder.Set(b, "From", sq.Alias(from, alias)).(SelectBuilder)
}

// JoinClause adds a join clause to the query.
func (b SelectBuilder) JoinClause(pred interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Joins", newPart(pred, args...)).(SelectBuilder)
}

// Join adds a JOIN clause to the query.
func (b SelectBuilder) Join(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("JOIN "+join, rest...)
}

// LeftJoin adds a LEFT JOIN clause to the query.
func (b SelectBuilder) LeftJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("LEFT JOIN "+join, rest...)
}

// RightJoin adds a RIGHT JOIN clause to the query.
func (b SelectBuilder) RightJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("RIGHT JOIN "+join, rest...)
}

// InnerJoin adds a INNER JOIN clause to the query.
func (b SelectBuilder) InnerJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("INNER JOIN "+join, rest...)
}

// CrossJoin adds a CROSS JOIN clause to the query.
func (b SelectBuilder) CrossJoin(join string, rest ...interface{}) SelectBuilder {
	return b.JoinClause("CROSS JOIN "+join, rest...)
}

// Where adds an expression to the WHERE clause of the query.
//
// Expressions are ANDed together in the generated SQL.
//
// Where accepts several types for its pred argument:
//
// nil OR "" - ignored.
//
// string - SQL expression.
// If the expression has SQL placeholders then a set of arguments must be passed
// as well, one for each placeholder.
//
// map[string]interface{} OR Eq - map of SQL expressions to values. Each key is
// transformed into an expression like "<key> = ?", with the corresponding value
// bound to the placeholder. If the value is nil, the expression will be "<key>
// IS NULL". If the value is an array or slice, the expression will be "<key> IN
// (?,?,...)", with one placeholder for each item in the value. These expressions
// are ANDed together.
//
// Where will panic if pred isn't any of the above types.
func (b SelectBuilder) Where(pred interface{}, args ...interface{}) SelectBuilder {
	if pred == nil || pred == "" {
		return b
	}
	return builder.Append(b, "WhereParts", newWherePart(pred, args...)).(SelectBuilder)
}

// GroupBy adds GROUP BY expressions to the query.
func (b SelectBuilder) GroupBy(groupBys ...string) SelectBuilder {
	return builder.Extend(b, "GroupBys", groupBys).(SelectBuilder)
}

// Having adds an expression to the HAVING clause of the query.
//
// See Where.
func (b SelectBuilder) Having(pred interface{}, rest ...interface{}) SelectBuilder {
	return builder.Append(b, "HavingParts", newWherePart(pred, rest...)).(SelectBuilder)
}

// OrderByClause adds ORDER BY clause to the query.
func (b SelectBuilder) OrderByClause(pred interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "OrderByParts", newPart(pred, args...)).(SelectBuilder)
}

// OrderBy adds ORDER BY expressions to the query.
func (b SelectBuilder) OrderBy(orderBys ...string) SelectBuilder {
	for _, orderBy := range orderBys {
		b = b.OrderByClause(orderBy)
	}

	return b
}

// Limit sets a LIMIT clause on the query.
func (b SelectBuilder) Limit(limit uint64) SelectBuilder {
	return builder.Set(b, "Limit", fmt.Sprintf("%d", limit)).(SelectBuilder)
}

// RemoveLimit ALL allows accessing all records with limit
func (b SelectBuilder) RemoveLimit() SelectBuilder {
	return builder.Delete(b, "Limit").(SelectBuilder)
}

// Offset sets a OFFSET clause on the query.
func (b SelectBuilder) Offset(offset uint64) SelectBuilder {
	return builder.Set(b, "Offset", fmt.Sprintf("%d", offset)).(SelectBuilder)
}

// RemoveOffset removes OFFSET clause.
func (b SelectBuilder) RemoveOffset() SelectBuilder {
	return builder.Delete(b, "Offset").(SelectBuilder)
}

// Suffix adds an expression to the end of the query
func (b SelectBuilder) Suffix(sql string, args ...interface{}) SelectBuilder {
	return b.SuffixExpr(sq.Expr(sql, args...))
}

// SuffixExpr adds an expression to the end of the query
func (b SelectBuilder) SuffixExpr(expr Sqlizer) SelectBuilder {
	return builder.Append(b, "Suffixes", expr).(SelectBuilder)
}

// Union adds a UNION clause to the query
func (b SelectBuilder) Union(query interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "Union", newUnionPart(query, args...)).(SelectBuilder)
}

// UnionAll adds a UNION ALL clause to the query
func (b SelectBuilder) UnionAll(query interface{}, args ...interface{}) SelectBuilder {
	return builder.Append(b, "UnionAll", newUnionPart(query, args...)).(SelectBuilder)
}
