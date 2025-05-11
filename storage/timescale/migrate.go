package timescale

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"strings"
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
				return fmt.Errorf("failed to read migration sql :: err=%w", err)
			}
			sqlToRun := strings.Replace(string(sqlStmt), "\n", "", -1)
			_, err = pool.Exec(ctx, sqlToRun)

			if err != nil {
				return fmt.Errorf("failed to execute migration sql :: err=%w", err)
			}
		}
		return nil
	})

	return err
}
