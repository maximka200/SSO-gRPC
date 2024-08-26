package auth

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/services/auth"

	ssov1 "github.com/maximka200/buffpr/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int64) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, userID int64) (flag bool, err error)
	CreateApp(ctx context.Context, name string, secret string) (appId int64, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func RegisterServ(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is empty")
	}
	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "App_id is empty")
	}
	token, err := s.auth.Login(ctx, req.Email, req.Password, int64(req.AppId))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.NotFound, "Invalid credentials")
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	panic("Logout not implemented")
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "User id is empty")
	}

	IsAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("User not found with id: %d", req.GetUserId()))
		}
		return nil, status.Error(codes.Internal, "Not correctly user_id or internal error")
	}

	return &ssov1.IsAdminResponse{IsAdmin: IsAdmin}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Password is empty")
	}
	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("User already exist with email: %s", req.GetEmail()))
		}
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverAPI) CreateApp(ctx context.Context, req *ssov1.CreateAppRequest) (*ssov1.CreateAppResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Name is empty")
	}
	if req.GetSecret() == "" {
		return nil, status.Error(codes.InvalidArgument, "Secret is empty")
	}
	appId, err := s.auth.CreateApp(ctx, req.Name, req.Secret)
	if err != nil {
		if errors.Is(err, auth.ErrAppExist) {
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("App already exist with email: %s", req.GetName()))
		}
		return nil, status.Errorf(codes.Internal, "Iternal error: %s", err)
	}
	return &ssov1.CreateAppResponse{AppId: appId}, nil
}

// implement delete user from db
