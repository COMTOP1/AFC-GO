// Package svcerr defines the typed errors services return, so the JSON API
// can map them to responses consistently.
package svcerr

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Kind classifies a service error.
type Kind int

const (
	KindNotFound Kind = iota + 1
	KindForbidden
	KindInvalid
	KindConflict
)

// Error is a service error that transports translate into a response.
type Error struct {
	Kind    Kind
	Message string
	Fields  map[string]string
	Err     error
}

func (e *Error) Error() string {
	msg := e.Message
	if len(e.Fields) > 0 {
		keys := make([]string, 0, len(e.Fields))
		for k := range e.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, k+": "+e.Fields[k])
		}
		msg += ": " + strings.Join(parts, "; ")
	}
	if e.Err != nil {
		msg += ": " + e.Err.Error()
	}
	return msg
}

func (e *Error) Unwrap() error { return e.Err }

func NotFound(message string, err error) *Error {
	return &Error{Kind: KindNotFound, Message: message, Err: err}
}

func Forbidden(message string) *Error {
	return &Error{Kind: KindForbidden, Message: message}
}

func Conflict(message string, err error) *Error {
	return &Error{Kind: KindConflict, Message: message, Err: err}
}

func Invalid(fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Message: "validation failed", Fields: fields}
}

func InvalidField(field, message string) *Error {
	return Invalid(map[string]string{field: message})
}

// Fields accumulates validation messages; the first message per field wins.
type Fields map[string]string

func (f Fields) Add(field, message string) {
	if _, ok := f[field]; !ok {
		f[field] = message
	}
}

// Err returns an Invalid error, or nil when no field failed.
func (f Fields) Err() error {
	if len(f) == 0 {
		return nil
	}
	return Invalid(f)
}

// FromStore turns a store error into a service error: sql.ErrNoRows
// becomes NotFound, anything else is wrapped unchanged.
func FromStore(err error, what string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return NotFound(what+" not found", err)
	}
	return fmt.Errorf("%s: %w", what, err)
}

// As reports whether err is (or wraps) a service error.
func As(err error) (*Error, bool) {
	var e *Error
	ok := errors.As(err, &e)
	return e, ok
}
