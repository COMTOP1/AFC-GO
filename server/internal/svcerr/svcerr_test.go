package svcerr_test

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/svcerr"
)

func TestFromStoreMapsNoRowsToNotFound(t *testing.T) {
	err := svcerr.FromStore(fmt.Errorf("failed to get news article: %w", sql.ErrNoRows), "news article")
	se, ok := svcerr.As(err)
	require.True(t, ok)
	assert.Equal(t, svcerr.KindNotFound, se.Kind)
	assert.Equal(t, "news article not found", se.Message)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestFromStoreWrapsOtherErrors(t *testing.T) {
	boom := errors.New("connection refused")
	err := svcerr.FromStore(boom, "news article")
	_, ok := svcerr.As(err)
	assert.False(t, ok)
	require.ErrorIs(t, err, boom)
	assert.NoError(t, svcerr.FromStore(nil, "x"))
}

func TestInvalidErrorListsFieldsSorted(t *testing.T) {
	err := svcerr.Invalid(map[string]string{"title": "title is required", "file": "unsupported file type: text/html"})
	assert.Equal(t, "validation failed: file: unsupported file type: text/html; title: title is required", err.Error())
}

func TestFieldsErr(t *testing.T) {
	f := svcerr.Fields{}
	require.NoError(t, f.Err())
	f.Add("name", "name is required")
	f.Add("name", "second message is ignored")
	se, ok := svcerr.As(f.Err())
	require.True(t, ok)
	assert.Equal(t, map[string]string{"name": "name is required"}, se.Fields)
}
