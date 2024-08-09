package app

import (
	"fmt"
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"time"
)

type App struct {
	gRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App { // TTL - time to live

	// init storage
	// init auth service

	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		gRPCSrv: grpcApp,
	}
}

func (app *App) MustRun() {
	if err := app.gRPCSrv.Run(); err != nil {
		err = fmt.Errorf("error run server: %w", err)
		panic(err)
	}
}
