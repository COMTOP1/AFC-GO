package filecheck

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/COMTOP1/AFC-GO/server/internal/testdb"
)

type fakeObjects struct {
	mu      sync.Mutex
	keys    map[string]bool
	deleted []string
}

func newObjects(keys ...string) *fakeObjects {
	o := &fakeObjects{keys: map[string]bool{}}
	for _, k := range keys {
		o.keys[k] = true
	}
	return o
}

func (o *fakeObjects) List(context.Context) ([]string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return slices.Sorted(maps.Keys(o.keys)), nil
}

func (o *fakeObjects) Delete(_ context.Context, key string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.keys, key)
	o.deleted = append(o.deleted, key)
	return nil
}

func byKey(ds []Discrepancy) map[string]Discrepancy {
	out := map[string]Discrepancy{}
	for _, d := range ds {
		out[string(d.Kind)+":"+d.Key] = d
	}
	return out
}

func TestCheckRecordsWaitsThenDeletes(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	store := NewStore(db)
	refs, err := store.References(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, refs, "the seed data has files")

	// The bucket holds every referenced file but one, plus an orphan and a
	// file only an article links to.
	var keys []string
	var missingKey string
	for k := range refs {
		if missingKey == "" {
			missingKey = k
			continue
		}
		keys = append(keys, k)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO news (title, content, date) VALUES ('Linked', '<img src="https://cdn.test/bucket/news/linked%20pic.jpg">', now())`)
	require.NoError(t, err)
	objects := newObjects(append(keys, "orphan.png", "news/linked pic.jpg")...)

	start := time.Date(2026, 10, 16, 3, 0, 0, 0, time.UTC)
	clock := start
	checker := NewChecker(store, objects)
	checker.now = func() time.Time { return clock }

	report, err := checker.Run(ctx, true)
	require.NoError(t, err)
	found := byKey(report.Differences)
	require.Len(t, found, 2, "%v", found)
	assert.Contains(t, found, "unreferenced:orphan.png")
	assert.NotContains(t, found, "unreferenced:news/linked pic.jpg", "a file linked from an article is in use")
	require.Contains(t, found, "missing:"+missingKey)
	assert.Equal(t, refs[missingKey][0], found["missing:"+missingKey].Detail, "says which row points at it")
	assert.Empty(t, report.Deleted, "nothing is deleted on first sight")

	clock = start.Add(29 * 24 * time.Hour)
	report, err = checker.Run(ctx, true)
	require.NoError(t, err)
	assert.Empty(t, report.Deleted, "29 days is not long enough")
	assert.Equal(t, start, byKey(report.Differences)["unreferenced:orphan.png"].FirstSeen.UTC(), "first sight is kept")

	clock = start.Add(30 * 24 * time.Hour)
	report, err = checker.Run(ctx, true)
	require.NoError(t, err)
	assert.Equal(t, []string{"orphan.png"}, report.Deleted)
	assert.Equal(t, []string{"orphan.png"}, objects.deleted)
	assert.NotContains(t, byKey(report.Differences), "unreferenced:orphan.png")
	assert.Contains(t, byKey(report.Differences), "missing:"+missingKey, "missing files are only reported")

	// Once the missing file turns up, its row goes.
	objects.keys[missingKey] = true
	report, err = checker.Run(ctx, true)
	require.NoError(t, err)
	assert.Empty(t, report.Differences)
}

func TestCheckWithoutDeleteNeverDeletes(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	objects := newObjects("orphan.png")
	checker := NewChecker(NewStore(db), objects)
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	checker.now = func() time.Time { return clock }
	_, err := checker.Run(ctx, false)
	require.NoError(t, err)
	clock = clock.Add(365 * 24 * time.Hour)
	_, err = checker.Run(ctx, false)
	require.NoError(t, err)
	assert.Empty(t, objects.deleted)
}

func TestCheckRefusesToDeleteAFloodOfFiles(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	keys := make([]string, 0, 30)
	for i := range 30 {
		keys = append(keys, fmt.Sprintf("stray/%02d.png", i))
	}
	objects := newObjects(keys...)
	checker := NewChecker(NewStore(db), objects)
	clock := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	checker.now = func() time.Time { return clock }
	_, err := checker.Run(ctx, true)
	require.NoError(t, err)
	clock = clock.Add(31 * 24 * time.Hour)
	report, err := checker.Run(ctx, true)
	require.NoError(t, err)
	assert.Empty(t, objects.deleted)
	assert.Contains(t, report.DeletionSkipped, "30 of 30 files look unreferenced")
}

func TestWithLockLetsOneInstanceRun(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	store := NewStore(db)
	ran, err := store.WithLock(ctx, func(ctx context.Context) error {
		inner, innerErr := store.WithLock(ctx, func(context.Context) error {
			t.Fatal("a second instance must not run while the first holds the lock")
			return nil
		})
		require.NoError(t, innerErr)
		assert.False(t, inner)
		return nil
	})
	require.NoError(t, err)
	assert.True(t, ran)

	ran, err = store.WithLock(ctx, func(context.Context) error { return nil })
	require.NoError(t, err)
	assert.True(t, ran, "the lock is released afterwards")
}

func TestMentioned(t *testing.T) {
	texts := []string{`<a href="https://cdn.test/bucket/document/club%20rules.pdf">rules</a>`, "see news/abc.jpg"}
	assert.True(t, mentioned(texts, "document/club rules.pdf"))
	assert.True(t, mentioned(texts, "news/abc.jpg"))
	assert.False(t, mentioned(texts, "news/other.jpg"))
}

func TestCheckFindsFilesInTablesTheCodeDoesntKnow(t *testing.T) {
	db, _ := testdb.Open(t)
	ctx := context.Background()
	// A table from the old site, and one with no id column.
	_, err := db.ExecContext(ctx, `CREATE TABLE legacy_gallery (id SERIAL PRIMARY KEY, file_name TEXT);
		INSERT INTO legacy_gallery (file_name) VALUES ('old-photo.jpg');
		CREATE TABLE attachments (file_name VARCHAR(200));
		INSERT INTO attachments VALUES ('attach.pdf');
		UPDATE teams SET description = 'Kit list: https://cdn.test/bucket/team/kit%20list.pdf' WHERE id = (SELECT min(id) FROM teams)`)
	require.NoError(t, err)

	store := NewStore(db)
	refs, err := store.References(ctx)
	require.NoError(t, err)
	assert.Contains(t, refs["old-photo.jpg"][0], "legacy_gallery ")
	assert.Equal(t, []string{"attachments"}, refs["attach.pdf"])

	objects := newObjects("old-photo.jpg", "attach.pdf", "team/kit list.pdf", "orphan.png")
	for k := range refs {
		objects.keys[k] = true
	}
	report, err := NewChecker(store, objects).Run(ctx, false)
	require.NoError(t, err)
	found := byKey(report.Differences)
	assert.Equal(t, map[string]Discrepancy{"unreferenced:orphan.png": found["unreferenced:orphan.png"]}, found,
		"only the orphan: legacy tables and links in any text column count as references")
}
