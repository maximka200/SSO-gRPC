package app

import (
	"fmt"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	sqlite "sso/internal/storage/sqllite"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App { // TTL - time to live

	storage, err := sqlite.NewStorage(storagePath)
	if err != nil {
		panic(err)
	}

	auth := auth.NewAuth(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, grpcPort, auth)

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
