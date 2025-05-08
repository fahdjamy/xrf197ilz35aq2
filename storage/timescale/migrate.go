package timescale

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"path/filepath"
	"xrf197ilz35aq2/internal"
)

func MigrateTimescaleTables(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsRelativePath := "storage/timescale/migrations"

	dir, err := internal.IsDir(migrationsRelativePath)
	if err != nil {
		return err
	}
	if !dir {
		return fmt.Errorf("migrations path not a directory")
	}

	err = filepath.Walk(migrationsRelativePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			sqlStmt, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Error reading file %s: %v\n", path, err)
				return err
			}
			_, err = pool.Exec(ctx, string(sqlStmt))
			if err != nil {
				slog.Warn("Error executing migration sql", "path", path, "err", err)
			}
		}
		return nil
	})

	return nil
}
