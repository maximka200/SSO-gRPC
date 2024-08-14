package auth

import (
	"context"

	ssov1 "github.com/GolangLessons/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const emptyValue = 0

type Auth interface {
	Login(ctx context.Context, email string, password string, appId int32) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	isAdmin(ctx context.Context, userID int64) (flag bool, err error)
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
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}
	if req.GetAppId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "app_id is empty")
	}
	// todo: impl Login
	token, err := s.auth.Login(ctx, req.Email, req.Password, req.AppId)
	if err != nil {
		// todo: различитель not correct password/login & iternal err
		return nil, status.Error(codes.Internal, "Not correctly password/login or internal error")
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	panic("Logout not implemented")
}

func (s *serverAPI) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if req.GetUserId() == emptyValue {
		return nil, status.Error(codes.InvalidArgument, "user_id is empty")
	}

	IsAdmin, err := s.auth.isAdmin(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "Not correctly user_id or internal error")
	}

	return &ssov1.IsAdminResponse{IsAdmin: IsAdmin}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is empty")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is empty")
	}
	userId, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		// handling зфыы
		return nil, status.Error(codes.Internal, "Already regist or internal error")
	}
	return &ssov1.RegisterResponse{UserId: userId}, nil
}
