// Package squirrel provides a fluent SQL generator.
//
// See https://github.com/Masterminds/squirrel for examples.
package squirrel

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

// Sqlizer is the interface that wraps the ToSql method.
//
// ToSql returns an SQL representation of the Sqlizer, along with a slice of args
// as passed to e.g. database/sql.Exec. It can also return an error.
type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

// rawSqlizer is expected to do what Sqlizer does, but without finalising placeholders.
// This is useful for nested queries.
type rawSqlizer interface {
	toSqlRaw() (string, []interface{}, error)
}

// Execer is the interface that wraps the Exec method.
//
// Exec executes the given query as implemented by database/sql.Exec.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Queryer is the interface that wraps the Query method.
//
// Query executes the given query as implemented by database/sql.Query.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// QueryRower is the interface that wraps the QueryRow method.
//
// QueryRow executes the given query as implemented by database/sql.QueryRow.
type QueryRower interface {
	QueryRow(query string, args ...interface{}) RowScanner
}

// BaseRunner groups the Execer and Queryer interfaces.
type BaseRunner interface {
	Execer
	Queryer
}

// Runner groups the Execer, Queryer, and QueryRower interfaces.
type Runner interface {
	Execer
	Queryer
	QueryRower
}

// WrapStdSql wraps a type implementing the standard SQL interface with methods that
// squirrel expects.
//
//nolint:revive
func WrapStdSql(stdSql StdSql) Runner {
	return &stdsqlRunner{stdSql}
}

// StdSql encompasses the standard methods of the *sql.DB type, and other types that
// wrap these methods.
//
//nolint:revive
type StdSql interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	Exec(string, ...interface{}) (sql.Result, error)
}

type stdsqlRunner struct {
	StdSql
}

func (r *stdsqlRunner) QueryRow(query string, args ...interface{}) RowScanner {
	return r.StdSql.QueryRow(query, args...)
}

func setRunWith(b interface{}, runner BaseRunner) interface{} {
	switch r := runner.(type) {
	case sq.StdSqlCtx:
		runner = sq.WrapStdSqlCtx(r)
	case StdSql:
		runner = WrapStdSql(r)
	}
	return builder.Set(b, "RunWith", runner)
}
