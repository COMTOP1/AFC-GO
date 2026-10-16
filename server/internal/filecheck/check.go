// Package filecheck compares the database's file references with the file
// store, records the differences, and deletes files nothing has referenced
// for 30 days.
package filecheck

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("github.com/COMTOP1/AFC-GO/server/internal/filecheck")

// MaxRetention is the longest a file may go unreferenced before it is
// deleted, and the default (FILE_RETENTION can make it shorter, e.g. on dev).
const MaxRetention = 30 * 24 * time.Hour

// ParseRetention reads a retention period: whole days ("1d", "7d") or a Go
// duration ("12h", "90m"). "" means MaxRetention. It must be above zero and
// at most MaxRetention.
func ParseRetention(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return MaxRetention, nil
	}
	var d time.Duration
	if days, ok := strings.CutSuffix(raw, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil {
			return 0, fmt.Errorf("invalid retention %q: want e.g. 7d or 12h", raw)
		}
		d = time.Duration(n) * 24 * time.Hour
	} else {
		var err error
		if d, err = time.ParseDuration(raw); err != nil {
			return 0, fmt.Errorf("invalid retention %q: want e.g. 7d or 12h", raw)
		}
	}
	if d <= 0 || d > MaxRetention {
		return 0, fmt.Errorf("retention %q must be above zero and no more than 30d", raw)
	}
	return d, nil
}

// Objects is the file store (satisfied by *storage.Store).
type Objects interface {
	List(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, key string) error
}

// Checker runs file checks.
type Checker struct {
	store     *Store
	objects   Objects
	now       func() time.Time
	retention time.Duration
}

// NewChecker deletes files after MaxRetention unless WithRetention says
// otherwise.
func NewChecker(store *Store, objects Objects) *Checker {
	return &Checker{store: store, objects: objects, now: time.Now, retention: MaxRetention}
}

// WithRetention sets how long a file may go unreferenced before it is
// deleted (see ParseRetention); it is capped at MaxRetention.
func (c *Checker) WithRetention(d time.Duration) *Checker {
	if d > 0 && d <= MaxRetention {
		c.retention = d
	}
	return c
}

// Report is what one check found and did.
type Report struct {
	Objects    int
	References int
	// Retention is how long a file may go unreferenced before it is deleted.
	Retention time.Duration
	// Differences are all recorded differences after the check, oldest first.
	Differences []Discrepancy
	// Deleted are files removed for being unreferenced for 30 days.
	Deleted []string
	// DeletionSkipped explains why nothing was deleted, when deletion was asked
	// for but held back.
	DeletionSkipped string
}

// Run compares the database with the file store and records the
// differences. With deleteDue it also deletes files that have been
// unreferenced for the retention period.
func (c *Checker) Run(ctx context.Context, deleteDue bool) (Report, error) {
	ctx, span := tracer.Start(ctx, "filecheck.Run")
	defer span.End()
	now := c.now().UTC()

	keys, err := c.objects.List(ctx)
	if err != nil {
		return Report{}, err
	}
	refs, err := c.store.References(ctx)
	if err != nil {
		return Report{}, err
	}
	texts, err := c.store.Texts(ctx)
	if err != nil {
		return Report{}, err
	}

	inBucket := make(map[string]bool, len(keys))
	var found []Discrepancy
	unreferenced := map[string]bool{}
	for _, key := range keys {
		inBucket[key] = true
		if len(refs[key]) == 0 && !mentioned(texts, key) {
			found = append(found, Discrepancy{Key: key, Kind: Unreferenced})
			unreferenced[key] = true
		}
	}
	for key, rows := range refs {
		if !inBucket[key] {
			slices.Sort(rows)
			found = append(found, Discrepancy{Key: key, Kind: Missing, Detail: strings.Join(rows, ", ")})
		}
	}
	if err = c.store.Record(ctx, now, found); err != nil {
		return Report{}, err
	}

	report := Report{Objects: len(keys), References: len(refs), Retention: c.retention}
	if deleteDue {
		report.Deleted, report.DeletionSkipped, err = c.deleteDue(ctx, now, len(keys), unreferenced)
		if err != nil {
			return report, err
		}
	}
	if report.Differences, err = c.store.List(ctx); err != nil {
		return report, err
	}
	return report, nil
}

// deleteDue deletes files recorded as unreferenced for the retention period
// that this check still found unreferenced.
func (c *Checker) deleteDue(ctx context.Context, now time.Time, objects int, unreferenced map[string]bool) ([]string, string, error) {
	// A sudden flood of unreferenced files means something is wrong (the
	// wrong bucket, a database read that came back empty), not that the
	// club stopped using its files: refuse to delete anything.
	if len(unreferenced) > 20 && len(unreferenced)*2 > objects {
		reason := fmt.Sprintf("%d of %d files look unreferenced; not deleting anything until that is checked", len(unreferenced), objects)
		slog.ErrorContext(ctx, "file check: "+reason)
		return nil, reason, nil
	}
	recorded, err := c.store.List(ctx)
	if err != nil {
		return nil, "", err
	}
	var deleted []string
	for _, d := range recorded {
		if d.Kind != Unreferenced || !unreferenced[d.Key] || now.Sub(d.FirstSeen) < c.retention {
			continue
		}
		if err = c.objects.Delete(ctx, d.Key); err != nil {
			return deleted, "", err
		}
		if err = c.store.Forget(ctx, d.Key, Unreferenced); err != nil {
			return deleted, "", err
		}
		deleted = append(deleted, d.Key)
	}
	return deleted, "", nil
}

// mentioned reports whether any stored text links key directly, as a raw
// key or as it appears in a file URL (path segments escaped).
func mentioned(texts []string, key string) bool {
	segments := strings.Split(key, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	escaped := strings.Join(segments, "/")
	for _, t := range texts {
		if strings.Contains(t, key) || strings.Contains(t, escaped) {
			return true
		}
	}
	return false
}
