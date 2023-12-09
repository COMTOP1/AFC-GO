package squirrel

import (
	"bytes"
	"fmt"
	"strings"
)

// PlaceholderFormat is the interface that wraps the ReplacePlaceholders method.
//
// ReplacePlaceholders takes a SQL statement and replaces each question mark
// placeholder with a (possibly different) SQL placeholder.
type PlaceholderFormat interface {
	ReplacePlaceholders(sql string) (string, error)
}

var (
	// Dollar is a PlaceholderFormat instance that replaces placeholders with
	// dollar-prefixed positional placeholders (e.g. $1, $2, $3).
	Dollar = dollarFormat{}
)

type dollarFormat struct{}

func (dollarFormat) ReplacePlaceholders(sql string) (string, error) {
	return replacePositionalPlaceholders(sql, "$")
}

func replacePositionalPlaceholders(sql, prefix string) (string, error) {
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(sql, "?")
		if p == -1 {
			break
		}

		if len(sql[p:]) > 1 && sql[p:p+2] == "??" { // escape ?? => ?
			buf.WriteString(sql[:p])
			buf.WriteString("?")
			if len(sql[p:]) == 1 {
				break
			}
			sql = sql[p+2:]
		} else {
			i++
			buf.WriteString(sql[:p])
			_, _ = fmt.Fprintf(buf, "%s%d", prefix, i)
			sql = sql[p+1:]
		}
	}

	buf.WriteString(sql)
	return buf.String(), nil
}
