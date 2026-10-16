package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/filecheck"
)

func TestPrintReport(t *testing.T) {
	seen := time.Date(2026, 10, 16, 3, 0, 0, 0, time.UTC)
	var out bytes.Buffer
	printReport(&out, filecheck.Report{
		Objects: 3, References: 2, Retention: filecheck.MaxRetention,
		Differences: []filecheck.Discrepancy{
			{Key: "orphan.png", Kind: filecheck.Unreferenced, FirstSeen: seen},
			{Key: "user/gone.png", Kind: filecheck.Missing, Detail: "users 4", FirstSeen: seen},
		},
		Deleted: []string{"old.png"},
	})
	for _, want := range []string{
		"3 files in the bucket, 2 referenced by the database",
		"orphan.png  first seen 16 Oct 2026, deleted from 15 Nov 2026 03:00",
		"user/gone.png  used by users 4",
		"Deleted (1):\n  old.png",
	} {
		assert.Contains(t, out.String(), want)
	}
}
