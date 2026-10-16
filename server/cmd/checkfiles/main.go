// Command checkfiles compares the database's file references with the file
// store and prints the differences: files nothing points to, and rows
// pointing at files that aren't there. It records them in
// file_discrepancies, like the server's daily check. With -delete it also
// deletes files unreferenced for 30 days; otherwise it changes no files.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // postgres driver

	"github.com/COMTOP1/AFC-GO/server/internal/filecheck"
	"github.com/COMTOP1/AFC-GO/server/internal/infrastructure/storage"
)

func main() {
	deleteDue := flag.Bool("delete", false, "also delete files unreferenced for 30 days")
	flag.Parse()
	_ = godotenv.Load(".env")
	_ = godotenv.Overload(".env.local")
	os.Exit(run(context.Background(), *deleteDue))
}

func run(ctx context.Context, deleteDue bool) int {
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		slog.Error(fmt.Sprintf("invalid option for DB_PORT: %+v", err))
		return 1
	}
	conn, err := sqlx.ConnectContext(ctx, "postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		os.Getenv("DB_HOSTNAME"), dbPort, os.Getenv("DB_USERNAME"), os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect to the database: %+v", err))
		return 1
	}
	defer conn.Close()
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}
	objects := storage.NewStore(ctx, storage.Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Region:    region,
		Bucket:    os.Getenv("S3_BUCKET"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	})

	store := filecheck.NewStore(conn)
	var report filecheck.Report
	ran, err := store.WithLock(ctx, func(ctx context.Context) error {
		var runErr error
		report, runErr = filecheck.NewChecker(store, objects).Run(ctx, deleteDue)
		return runErr
	})
	if err != nil {
		slog.Error(fmt.Sprintf("file check failed: %+v", err))
		return 1
	}
	if !ran {
		slog.Error("a server instance is running the file check right now; try again in a few minutes")
		return 1
	}
	printReport(os.Stdout, report)
	return 0
}

func printReport(w io.Writer, r filecheck.Report) {
	_, _ = fmt.Fprintf(w, "%d files in the bucket, %d referenced by the database\n\n", r.Objects, r.References) //nolint:errcheck // stdout
	var unreferenced, missing []filecheck.Discrepancy
	for _, d := range r.Differences {
		if d.Kind == filecheck.Missing {
			missing = append(missing, d)
		} else {
			unreferenced = append(unreferenced, d)
		}
	}
	_, _ = fmt.Fprintf(w, "Unreferenced files (%d): in the bucket, but nothing points to them\n", len(unreferenced)) //nolint:errcheck // stdout
	for _, d := range unreferenced {
		_, _ = fmt.Fprintf(w, "  %s  first seen %s, deleted from %s\n", d.Key, //nolint:errcheck // stdout
			d.FirstSeen.Format("2 Jan 2006"), d.FirstSeen.Add(filecheck.RetainUnreferenced).Format("2 Jan 2006"))
	}
	_, _ = fmt.Fprintf(w, "\nMissing files (%d): the database points to them, but they aren't in the bucket\n", len(missing)) //nolint:errcheck // stdout
	for _, d := range missing {
		_, _ = fmt.Fprintf(w, "  %s  used by %s\n", d.Key, d.Detail) //nolint:errcheck // stdout
	}
	if len(r.Deleted) > 0 {
		_, _ = fmt.Fprintf(w, "\nDeleted (%d):\n", len(r.Deleted)) //nolint:errcheck // stdout
		for _, k := range r.Deleted {
			_, _ = fmt.Fprintf(w, "  %s\n", k) //nolint:errcheck // stdout
		}
	}
	if r.DeletionSkipped != "" {
		_, _ = fmt.Fprintf(w, "\nNothing deleted: %s\n", r.DeletionSkipped) //nolint:errcheck // stdout
	}
}
