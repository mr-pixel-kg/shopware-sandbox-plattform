package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/database"
	"github.com/pressly/goose/v3"
)

const migrationDir = "internal/database/migrations"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command (expected: up, down, status, version, reset, fresh, create)")
	}

	command := args[0]
	if command == "create" {
		if len(args) < 2 || args[1] == "" {
			return fmt.Errorf("missing migration name")
		}
		goose.SetSequential(true)
		return goose.Create(nil, migrationDir, args[1], "sql")
	}

	cfg := config.MustLoad()
	gormDB, err := database.Connect(cfg.Database, 0)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("open sql database handle: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	switch command {
	case "up":
		return goose.Up(sqlDB, migrationDir)
	case "down":
		return goose.Down(sqlDB, migrationDir)
	case "status":
		return goose.Status(sqlDB, migrationDir)
	case "version":
		return goose.Version(sqlDB, migrationDir)
	case "reset":
		return goose.Reset(sqlDB, migrationDir)
	case "fresh":
		if err := resetPublicSchema(sqlDB); err != nil {
			return err
		}
		return goose.Up(sqlDB, migrationDir)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func resetPublicSchema(sqlDB *sql.DB) error {
	statements := []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
		"GRANT ALL ON SCHEMA public TO CURRENT_USER",
	}

	for _, stmt := range statements {
		if _, err := sqlDB.Exec(stmt); err != nil {
			return fmt.Errorf("reset public schema with %q: %w", stmt, err)
		}
	}

	return nil
}
