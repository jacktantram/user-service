package transportgrpc

import (
	"context"
	"errors"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/jacktantram/user-service/services/user-service/internal/domain"
	uuid "github.com/kevinburke/go.uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) CreateUser(ctx context.Context, request *userServiceV1.CreateUserRequest) (*userServiceV1.CreateUserResponse, error) {
	u := &domain.User{}
	u.FromProto(request.GetUser())
	if err := s.validate.Struct(u); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	if err := s.service.CreateUser(ctx, request.GetUser()); err != nil {
		if errors.Is(err, domain.ErrCreateUserEmailUnique) {
			return nil, status.New(codes.AlreadyExists, "user already exists with this email").Err()
		}
		log.WithError(err).Error("unable to create user")
		return nil, errSomethingWentWrong
	}

	log.WithContext(ctx).WithFields(log.Fields{
		"user_id": request.User.Id,
	}).Info("user is created")

	return &userServiceV1.CreateUserResponse{User: request.User}, nil
}

func validateGetUser(req *userServiceV1.GetUserRequest) error {
	if req.Id == "" {
		return errors.New("user id must be provided")
	}
	if _, err := uuid.FromString(req.Id); err != nil {
		return errors.New("user id must be in the UUID format")
	}
	return nil
}

func (s *Server) GetUser(ctx context.Context, request *userServiceV1.GetUserRequest) (*userServiceV1.GetUserResponse, error) {
	if err := validateGetUser(request); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	user, err := s.service.GetUser(ctx, request.GetId())

	logger := log.WithFields(log.Fields{
		"user_id": request.Id,
	})

	if err != nil {
		if errors.Is(err, domain.ErrNoUser) {
			return nil, status.New(codes.NotFound, "user is not found").Err()
		}
		log.WithError(err).WithFields(log.Fields{"user_id": request.Id}).Error("unable to get user")
		return nil, errSomethingWentWrong
	}
	logger.WithContext(ctx).Info("user is fetched")
	return &userServiceV1.GetUserResponse{User: user}, nil
}

func (s *Server) ListUsers(ctx context.Context, request *userServiceV1.ListUsersRequest) (*userServiceV1.ListUsersResponse, error) {
	users, err := s.service.ListUsers(ctx, request.GetFilters(), request.Offset, request.Limit)
	if err != nil {
		log.WithError(err).WithFields(
			log.Fields{
				"filters": request.Filters, "offset": request.Offset,
				"limit": request.Limit,
			}).Error("unable to list users")

		return nil, errSomethingWentWrong
	}
	return &userServiceV1.ListUsersResponse{Users: users}, nil
}

func validateUpdateUser(req *userServiceV1.UpdateUserRequest) error {
	if req.User == nil {
		return errors.New("user must be provided")
	}
	if req.User.Id == "" {
		return errors.New("user id must be provided")
	}
	if len(req.UpdateFields) == 0 {
		return errors.New("at least one update field must be provided")
	}
	if len(req.UpdateFields) == 1 && req.UpdateFields[0] == v1.UpdateUserField_UPDATE_USER_FIELD_UNSPECIFIED {
		return errors.New("at least one update field must be provided and not be unspecified value")
	}
	updateFieldCount := make(map[v1.UpdateUserField]int, 0)
	for _, val := range req.UpdateFields {
		if _, ok := updateFieldCount[val]; ok {
			return errors.New("should only input unique update fields")
		} else {
			updateFieldCount[val] = 1
		}
	}
	return nil
}

func (s *Server) UpdateUser(ctx context.Context, request *userServiceV1.UpdateUserRequest) (*userServiceV1.UpdateUserResponse, error) {
	if err := validateUpdateUser(request); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}
	u := &domain.User{}
	u.FromProto(request.GetUser())
	if err := s.validate.Struct(u); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	logger := log.WithFields(log.Fields{
		"user_id":       request.User.Id,
		"update_fields": request.UpdateFields,
	})

	if err := s.service.UpdateUser(ctx, request.User, request.UpdateFields); err != nil {
		if errors.Is(err, domain.ErrNoUser) {
			return nil, status.New(codes.NotFound, err.Error()).Err()
		}
		logger.WithError(err).Error("unable to update user")
		return nil, errSomethingWentWrong
	}
	logger.WithContext(ctx).Info("user is deleted")
	return &userServiceV1.UpdateUserResponse{User: request.User}, nil
}

func validateDeleteUser(req *userServiceV1.DeleteUserRequest) error {
	if req.Id == "" {
		return errors.New("user id must be provided")
	}
	if _, err := uuid.FromString(req.Id); err != nil {
		return errors.New("user id must be in the UUID format")
	}
	return nil
}

func (s *Server) DeleteUser(ctx context.Context, request *userServiceV1.DeleteUserRequest) (*userServiceV1.DeleteUserResponse, error) {
	if err := validateDeleteUser(request); err != nil {
		return nil, status.New(codes.InvalidArgument, err.Error()).Err()
	}

	logger := log.WithFields(log.Fields{
		"user_id": request.Id,
	})
	if err := s.service.DeleteUser(ctx, request.Id); err != nil {
		if errors.Is(err, domain.ErrNoUser) {
			return nil, status.New(codes.NotFound, "user is not found").Err()
		}

		log.WithError(err).Error("unable to delete user")

		return nil, errSomethingWentWrong

	}
	logger.WithContext(ctx).Info("user is deleted")

	return &userServiceV1.DeleteUserResponse{}, nil
}
