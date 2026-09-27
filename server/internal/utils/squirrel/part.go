package squirrel

import (
	"fmt"
	"io"

	sq "github.com/Masterminds/squirrel"
)

type part struct {
	pred interface{}
	args []interface{}
}

func newPart(pred interface{}, args ...interface{}) sq.Sqlizer {
	return &part{pred, args}
}

//nolint:revive
func (p part) ToSql() (sql string, args []interface{}, err error) {
	switch pred := p.pred.(type) {
	case nil:
		// no-op
	case sq.Sqlizer:
		sql, args, err = nestedToSql(pred)
	case string:
		sql = pred
		args = p.args
	default:
		err = fmt.Errorf("expected string or Sqlizer, not %T", pred)
	}
	return
}

//nolint:revive
func nestedToSql(s sq.Sqlizer) (string, []interface{}, error) {
	if raw, ok := s.(rawSqlizer); ok {
		return raw.toSqlRaw()
	}
	return s.ToSql()
}

//nolint:revive
func appendToSql(parts []sq.Sqlizer, w io.Writer, sep string, args []interface{}) ([]interface{}, error) {
	for i, p := range parts {
		partSQL, partArgs, err := nestedToSql(p)
		if err != nil {
			return nil, err
		} else if len(partSQL) == 0 {
			continue
		}

		if i > 0 {
			_, err = io.WriteString(w, sep)
			if err != nil {
				return nil, err
			}
		}

		_, err = io.WriteString(w, partSQL)
		if err != nil {
			return nil, err
		}
		args = append(args, partArgs...)
	}
	return args, nil
}
