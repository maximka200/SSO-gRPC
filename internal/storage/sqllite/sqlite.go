package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/services/storage"

	"github.com/mattn/go-sqlite3"
)

const (
	usersTable = "users"
	appsTable  = "apps"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("couldn`t open db: %s", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("couldn`t connect db: %s", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx *context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "storage.sqlite.SaveUser"
	var id int64

	stmt, err := s.db.Prepare(fmt.Sprintf("INSERT INTO %s (email, password_hash) values ($1, $2)", usersTable))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(*ctx, email, passHash)
	if err != nil {
		var sqlliteErr sqlite3.Error

		if errors.As(err, &sqlliteErr) && sqlliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, storage.ErrUserExist
		}

		return 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"
	var us models.User

	stmt, err := s.db.Prepare(fmt.Sprintf("SELECT id, email, password_hash FROM %s WHERE email=$1", usersTable))
	if err != nil {
		return us, fmt.Errorf("%s: %s", op, err.Error())
	}

	result := stmt.QueryRowContext(ctx, email)

	if err = result.Scan(&us.ID, &us.Email, &us.PassHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return us, storage.ErrUserNotFound
		}
	}

	return us, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (models.App, error) {
	const op = "storage.sqlite.App"
	var app models.App

	stmt, err := s.db.Prepare(fmt.Sprintf("SELECT id, name, secret FROM %s WHERE id=$1", appsTable))
	if err != nil {
		return app, fmt.Errorf("%s: %s", op, err.Error())
	}

	result := stmt.QueryRowContext(ctx, appID)

	if err = result.Scan(&app.Id, &app.Name, &app.Secret); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app, storage.ErrAppNotFound
		}
	}

	return app, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"
	var res bool
	stmt, err := s.db.Prepare(fmt.Sprintf("SELECT is_admin FROM %s WHERE id=$1", usersTable))
	if err != nil {
		return false, fmt.Errorf("%s: %s", op, err.Error())
	}

	result := stmt.QueryRowContext(ctx, userID)

	if err = result.Scan(&res); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
	}

	return res, nil

}
