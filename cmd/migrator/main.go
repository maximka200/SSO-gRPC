package main

import (
	"errors"
	"flag"
	"fmt"

	// драйвера для работы с бд и файлами
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationPath, migrationTable string
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")
	flag.StringVar(&migrationPath, "migration-path", "", "path to migration")
	flag.StringVar(&migrationTable, "migration-table", "", "name of migration table") // для тестов
	flag.Parse()

	if storagePath == "" || migrationPath == "" {
		panic("migration or storage path is empty")
	}

	m, err := migrate.New("file://"+migrationPath, fmt.Sprintf("sqlite3://%s?x-migration-table=%s", storagePath, migrationTable)) // mb sourceUrl working only for unix system
	if err != nil {
		panic(fmt.Sprintf("error create migration: %s", err))
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migration to apply")
			return
		}

		panic(fmt.Sprintf("error start migration: %s", err))
	}

	fmt.Println("migrations applied successfully")
}
