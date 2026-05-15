package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := flag.String("dsn", os.Getenv("KEIYAKU_MYSQL_DSN"), "mysql dsn")
	dir := flag.String("dir", "migrations", "migration directory")
	direction := flag.String("direction", "up", "up or down")
	steps := flag.Int("steps", 1, "down migration steps")
	flag.Parse()
	if *dsn == "" {
		panic("dsn is required")
	}
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic(err)
	}
	if err := ensureMigrationTable(db); err != nil {
		panic(err)
	}
	switch *direction {
	case "up":
		err = migrateUp(db, *dir)
	case "down":
		err = migrateDown(db, *dir, *steps)
	default:
		err = fmt.Errorf("unsupported direction %q", *direction)
	}
	if err != nil {
		panic(err)
	}
}

func ensureMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
version VARCHAR(255) PRIMARY KEY,
applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	return err
}

func migrateUp(db *sql.DB, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		version := migrationVersion(file, ".up.sql")
		applied, err := migrationApplied(db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := execFile(db, file); err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)", version, time.Now().UTC()); err != nil {
			return err
		}
		fmt.Printf("applied %s\n", version)
	}
	return nil
}

func migrateDown(db *sql.DB, dir string, steps int) error {
	if steps <= 0 {
		return nil
	}
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version DESC LIMIT ?", steps)
	if err != nil {
		return err
	}
	defer rows.Close()
	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return err
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, version := range versions {
		file := filepath.Join(dir, version+".down.sql")
		if err := execFile(db, file); err != nil {
			return err
		}
		if _, err := db.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
			return err
		}
		fmt.Printf("reverted %s\n", version)
	}
	return nil
}

func migrationApplied(db *sql.DB, version string) (bool, error) {
	var found string
	err := db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&found)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func execFile(db *sql.DB, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sqlText := strings.TrimSpace(string(content))
	if sqlText == "" {
		return nil
	}
	for _, statement := range strings.Split(sqlText, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("execute %s: %w", file, err)
		}
	}
	return nil
}

func migrationVersion(file, suffix string) string {
	base := filepath.Base(file)
	return strings.TrimSuffix(base, suffix)
}
