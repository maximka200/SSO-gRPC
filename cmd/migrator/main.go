package main

import (
	_ "database/sql"
	"errors"
	"flag"
	"fmt"

	// драйвера для работы с бд и файлами
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	mustCreateMigrateUp()
}

/* change dirty values in db :
func updateDirty(val int, migrationsTable string) error {
	db, err := sql.Open("sqlite3", "internal/storage/sso.db")
	if err != nil {
		return fmt.Errorf("cannot open db: %s", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("UPDATE %s SET dirty=%d", migrationsTable, val))
	if err != nil {
		return fmt.Errorf("cannot execute the request to db: %s", err)
	}

	fmt.Println("dirty value changed successfully")

	return nil
}
*/

func mustCreateMigrateUp() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied")
}
