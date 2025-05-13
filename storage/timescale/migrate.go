package timescale

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"xrf197ilz35aq2/internal"
)

const (
	tsStatementsSep = "----"
)

func MigrateTimescaleTables(ctx context.Context, pool *pgxpool.Pool, logger slog.Logger) error {
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
			sqlStmtInFile := strings.Replace(string(sqlStmt), "\n", "", -1)
			statements := strings.Split(sqlStmtInFile, tsStatementsSep)
			if len(statements) != 2 {
				return fmt.Errorf("invalid migration scripts found in %s", path)
			}
			createTableStmt := statements[0]
			_, err = pool.Exec(ctx, createTableStmt)
			if err != nil {
				return fmt.Errorf("failed to execute migration sql :: err=%w", err)
			}

			createHypertableSQL := statements[1]
			_, err = pool.Exec(ctx, createHypertableSQL)
			if err != nil {
				return fmt.Errorf("failed create hypertable :: err=%w", err)
			}
			tableName := getTableName(createHypertableSQL)
			logger.Info("created table and hypertable", "tableName", tableName)
		}
		return nil
	})

	return err
}

func getTableName(createHypertableSQL string) string {
	parts := strings.Split(createHypertableSQL, "(")
	tableName := strings.Split(parts[1], ",")
	return tableName[0]
}
