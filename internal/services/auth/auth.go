package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	jwtlocal "sso/internal/lib"
	"sso/internal/services/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user exist")
	ErrInvalidAppID       = errors.New("invalid appID")
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(ctx *context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

// New returns a new object of the Auth struct
func NewAuth(log *slog.Logger, usrSaver UserSaver, usrProvider UserProvider, appProvider AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    usrSaver,
		usrProvider: usrProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appID int64) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("not corrected login/password")
			return "", fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user")
		return "", fmt.Errorf("%s:%w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("not corrected login/password")
		return "", fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	log.Info("successfully login user")

	token, err := jwtlocal.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("cannot generate token")
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(&ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			log.Error("user already exist")
			return 0, fmt.Errorf("%s:%w", op, ErrUserExists)
		}
		log.Error("failed to save user")
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("successfully register user")

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	const op = "auth.IsAdmin"

	log := slog.With(slog.String("op", op), slog.Int64("userId", userId))

	result, err := a.usrProvider.IsAdmin(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Error("user not found")
			return false, fmt.Errorf("%s:%w", op, ErrInvalidCredentials)
		}
		log.Error("error when searching for user")
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("success checking is admin", slog.Bool("isAdmin", result))

	return result, nil
}
