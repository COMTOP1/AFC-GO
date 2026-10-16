package filecheck

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job runs the file check, with deletion, once a day. Every server instance
// runs a Job; the Postgres lock means only one of them checks at a time.
type Job struct {
	checker  *Checker
	store    *Store
	interval time.Duration
	// firstRun is how long after Start the first check runs, so a restart
	// doesn't hit the bucket straight away.
	firstRun time.Duration

	stop     chan struct{}
	stopOnce sync.Once
}

func NewJob(checker *Checker, store *Store) *Job {
	return &Job{checker: checker, store: store, interval: 24 * time.Hour, firstRun: 10 * time.Minute, stop: make(chan struct{})}
}

// Start begins the daily checks until Stop.
func (j *Job) Start() {
	go func() { //nolint:gosec,contextcheck // background job outlives any request; context.Background() is intentional
		timer := time.NewTimer(j.firstRun)
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				j.RunOnce(context.Background())
				timer.Reset(j.interval)
			case <-j.stop:
				return
			}
		}
	}()
}

// Stop ends the job. It is safe to call more than once.
func (j *Job) Stop() {
	j.stopOnce.Do(func() { close(j.stop) })
}

// RunOnce runs one check with deletion, unless another instance is.
func (j *Job) RunOnce(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "filecheck.Job.RunOnce")
	defer span.End()
	var report Report
	ran, err := j.store.WithLock(ctx, func(ctx context.Context) error {
		var runErr error
		report, runErr = j.checker.Run(ctx, true)
		return runErr
	})
	switch {
	case err != nil:
		span.RecordError(err)
		slog.ErrorContext(ctx, "file check failed", "error", err)
	case !ran:
		slog.InfoContext(ctx, "file check skipped: another instance is running it")
	default:
		unreferenced, missing := 0, 0
		for _, d := range report.Differences {
			if d.Kind == Missing {
				missing++
			} else {
				unreferenced++
			}
		}
		slog.InfoContext(ctx, "file check done", "objects", report.Objects, "unreferenced", unreferenced,
			"missing", missing, "deleted", len(report.Deleted))
	}
}
