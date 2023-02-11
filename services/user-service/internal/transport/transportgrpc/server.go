package transportgrpc

//go:generate mockgen -source=server.go -destination=mocks/mock_service.go -package=mocks

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

var (
	errSomethingWentWrong = status.New(codes.Internal, "oops something went wrong!").Err()
)

// Service represents an interface for interacting with the main service.
type Service interface {
	CreateUser(ctx context.Context, user *v1.User) error
	GetUser(ctx context.Context, id string) (*v1.User, error)
	ListUsers(ctx context.Context, filters *userServiceV1.SelectUserFilters, offset uint64, limit uint64) ([]*v1.User, error)
	UpdateUser(ctx context.Context, userToUpdate *v1.User, updateFields []v1.UpdateUserField) error
	DeleteUser(ctx context.Context, id string) error
}

// Server defines a GRPC server
type Server struct {
	userServiceV1.UnimplementedUserServiceServer
	*grpc.Server
	service  Service
	validate *validator.Validate
}

// NewServer Creates a new server
func NewServer(server *grpc.Server, service Service) (*Server, error) {
	if (server == nil) || (service == nil) {
		return nil, errors.New("server and service must not be nil")
	}
	s := &Server{Server: server, service: service, validate: validator.New()}
	userServiceV1.RegisterUserServiceServer(server, s)
	return s, nil
}
