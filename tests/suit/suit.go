package suite

import (
	"context"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/maximka200/buffpr/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHost = "localhost"
)

type Suite struct {
	*testing.T // обьект для взаимодействия с тестами
	Cfg        *config.Config
	AuthClient ssov1.AuthClient
}

func NewSuite(t *testing.T) (context.Context, *Suite) {
	t.Helper()   // для пометки функции как вспомогательной
	t.Parallel() // параллельные тесты

	key := "../config/local.yaml"
	cfg := config.MustByLoad(key)

	// создаем контекст
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	// выполняется перед завершения теста
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	cc, err := grpc.DialContext(ctx, grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()), // говорим cерверу что будем использовать insecure соединение
	)
	if err != nil {
		t.Fatalf("grpc server connection error: %s", err)
	}

	return ctx, &Suite{T: t, Cfg: cfg, AuthClient: ssov1.NewAuthClient(cc)}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
