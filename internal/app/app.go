package app

import (
	"fmt"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage/postgresql"
	// sqlite "sso/internal/storage/sqllite"
	//"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App { // TTL - time to live

	storage, err := postgresql.NewDB(cfg)
	if err != nil {
		panic(err)
	}

	auth := auth.NewAuth(log, storage, storage, storage, storage, cfg.GRPC.Timeout)

	grpcApp := grpcapp.New(log, cfg.GRPC.Port, auth)

	return &App{
		GRPCSrv: grpcApp,
	}
}

func (app *App) MustRun() {
	if err := app.GRPCSrv.Run(); err != nil {
		err = fmt.Errorf("error run server: %w", err)
		panic(err)
	}
}
