package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/config"
	"sso/internal/domain/models"
	"sso/internal/services/storage"

	"github.com/lib/pq"
)

const (
	usersTable = "users"
	appsTable  = "apps"
)

type Storage struct {
	db *sql.DB
}

func NewDB(cfg *config.Config) (*Storage, error) {
	op := "storage.NewPostgreDB"

	db, err := sql.Open("postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.Username, cfg.DB.Password, cfg.DB.DBname, cfg.DB.SSLmode))
	if err != nil {
		return nil, fmt.Errorf("%s:%s", op, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DB.Timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s:%s", err, op)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "storage.postgres.SaveUser"

	stmt, err := s.db.Prepare(fmt.Sprintf("INSERT INTO %s (email, password_hash) values ($1, $2) RETURNING id", usersTable))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64

	err = stmt.QueryRowContext(ctx, email, passHash).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %s", op, err.Error())
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgresql.User"

	var us models.User

	stmt, err := s.db.Prepare(fmt.Sprintf("SELECT id, email, password_hash FROM %s WHERE email=$1", usersTable))
	if err != nil {
		return us, fmt.Errorf("%s: %s", op, err.Error())
	}

	result := stmt.QueryRowContext(ctx, email)

	if err = result.Scan(&us.ID, &us.Email, &us.PassHash); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return us, storage.ErrAppExist
			}
		}

		return us, fmt.Errorf("%s: %w", op, err)
	}

	return us, nil
}

func (s *Storage) App(ctx context.Context, appID int64) (models.App, error) {
	const op = "storage.postgresql.App"

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
	const op = "storage.postgresql.IsAdmin"

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

		return false, fmt.Errorf("%s: %s", op, err.Error())
	}

	return res, nil
}

func (s *Storage) SaveApp(ctx context.Context, name string, secret string) (int64, error) {
	const op = "storage.postgresql.CreateApp"

	stmt, err := s.db.Prepare(fmt.Sprintf("INSERT INTO %s (name, secret) values ($1, $2) RETURNING id", appsTable))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64
	if err := stmt.QueryRowContext(ctx, name, secret).Scan(&id); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return 0, storage.ErrAppExist
			}
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
