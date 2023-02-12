package transportgrpc_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/jacktantram/user-service/internal/domain"
	"github.com/jacktantram/user-service/internal/transport/transportgrpc"
	"github.com/jacktantram/user-service/internal/transport/transportgrpc/mocks"
	uuid "github.com/kevinburke/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestServer_CreateUser_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.CreateUserRequest
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockService, args args, want *userServiceV1.CreateUserResponse)
		args  args
		want  *userServiceV1.CreateUserResponse
	}{
		{
			name: "should be able to successfully create the user",
			args: args{request: &userServiceV1.CreateUserRequest{User: &v1.User{
				FirstName: "A name",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				UpdatedAt: nil,
			}}},
			setup: func(mockService *mocks.MockService, args args,
				want *userServiceV1.CreateUserResponse) {
				mockService.
					EXPECT().
					CreateUser(gomock.Any(), gomock.Eq(args.request.User)).
					DoAndReturn(func(ctx context.Context, user *v1.User) error {
						args.request.User = want.User
						return nil
					})
			},
			want: &userServiceV1.CreateUserResponse{
				User: &v1.User{
					Id:        "a-user",
					FirstName: "John",
					LastName:  "Gopher",
					Nickname:  "Goopher",
					Password:  "A-password",
					Email:     "jon@gopher.com",
					Country:   "GBR",
					CreatedAt: timestamppb.Now(),
					UpdatedAt: nil,
				}},
		}}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args, tt.want)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.CreateUser(context.Background(), tt.args.request)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServer_CreateUser_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.CreateUserRequest
	}
	tests := []struct {
		name            string
		setup           func(mockService *mocks.MockService, args args)
		args            args
		wantErrContains string
		wantStatusCode  codes.Code
	}{
		{
			name: "should return error unable to create a user due to invalid email",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "A name",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "anemailgopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Email' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing email",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "A name",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Email' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to invalid country code",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "A name",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "anemail@gopher.com",
					Country:   "DE",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Country' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing country",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "A name",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "a@email.com",
					Country:   "",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Country' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to invalid country code",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "A name",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "anemail@gopher.com",
					Country:   "DEUU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Country' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing first name",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "",
					LastName:  "Some field",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'FirstName' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing last name",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "Sopme",
					LastName:  "",
					Nickname:  "Nickanme",
					Password:  "a-password",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'LastName' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing nickname",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "Sopme",
					LastName:  "asdasd",
					Nickname:  "",
					Password:  "a-password",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Nickname' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due to missing password",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "Sopme",
					LastName:  "asdasd",
					Nickname:  "asdas",
					Password:  "",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup:           nil,
			wantErrContains: "Field validation for 'Password' failed",
			wantStatusCode:  codes.InvalidArgument,
		},
		{
			name: "should return error unable to create a user due service returning error",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "Sopme",
					LastName:  "asdasd",
					Nickname:  "asdas",
					Password:  "valid",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(errors.New("something bad"))
			},
			wantErrContains: "oops something went wrong!",
			wantStatusCode:  codes.Internal,
		},
		{
			name: "should return error due to email already existing",
			args: args{
				request: &userServiceV1.CreateUserRequest{User: &v1.User{
					FirstName: "Sopme",
					LastName:  "asdasd",
					Nickname:  "asdas",
					Password:  "valid",
					Email:     "anemail@gopher.com",
					Country:   "DEU",
					UpdatedAt: nil,
				}}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(domain.ErrCreateUserEmailUnique)
			},
			wantErrContains: "user already exists with this email",
			wantStatusCode:  codes.AlreadyExists,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.CreateUser(context.Background(), tt.args.request)
			assert.Nil(t, got)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrContains)
			assert.Equal(t, tt.wantStatusCode, status.Convert(err).Code())

		})
	}
}

func TestServer_GetUser_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.GetUserRequest
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockService, args args, want *userServiceV1.GetUserResponse)
		args  args
		want  *userServiceV1.GetUserResponse
	}{
		{
			name: "should be able to successfully get the user",
			args: args{request: &userServiceV1.GetUserRequest{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"}},
			setup: func(mockService *mocks.MockService, args args,
				want *userServiceV1.GetUserResponse) {
				mockService.EXPECT().GetUser(gomock.Any(), gomock.Eq(args.request.Id)).Return(want.User, nil)
			},
			want: &userServiceV1.GetUserResponse{User: &v1.User{
				Id:        "a8bdce5a-31dc-4647-98b5-ce9cb343138f",
				FirstName: "John",
				LastName:  "Gopher",
				Nickname:  "Goopher",
				Password:  "A-password",
				Email:     "jon@gopher.com",
				Country:   "GBR",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: nil,
			}},
		}}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args, tt.want)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.GetUser(context.Background(), tt.args.request)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServer_GetUser_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.GetUserRequest
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockService, args args)
		args    args
		wantErr error
	}{
		{
			name:    "should throw error if id is missing",
			args:    args{request: &userServiceV1.GetUserRequest{Id: ""}},
			setup:   nil,
			wantErr: status.Error(codes.InvalidArgument, "user id must be provided"),
		},
		{
			name:    "should throw error if id is not a uuid",
			args:    args{request: &userServiceV1.GetUserRequest{Id: "some-val"}},
			setup:   nil,
			wantErr: status.Error(codes.InvalidArgument, "user id must be in the UUID format"),
		},
		{
			name: "should return NotFound error if service returns domain.ErrNoUser",
			args: args{request: &userServiceV1.GetUserRequest{Id: uuid.NewV4().String()}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrNoUser)
			},
			wantErr: status.New(codes.NotFound, "user is not found").Err(),
		},
		{
			name: "should return internal error if service returns error",
			args: args{request: &userServiceV1.GetUserRequest{Id: uuid.NewV4().String()}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("some error"))
			},
			wantErr: status.New(codes.Internal, "oops something went wrong!").Err(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.GetUser(context.Background(), tt.args.request)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestServer_ListUsers_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.ListUsersRequest
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockService, args args, want *userServiceV1.ListUsersResponse)
		args  args
		want  *userServiceV1.ListUsersResponse
	}{
		{
			name: "should be able to successfully list users",
			args: args{
				request: &userServiceV1.ListUsersRequest{
					Filters: &userServiceV1.SelectUserFilters{Countries: []string{"GBR"}}, Offset: 0, Limit: 100},
			},
			setup: func(mockService *mocks.MockService, args args,
				want *userServiceV1.ListUsersResponse) {
				mockService.
					EXPECT().
					ListUsers(gomock.Any(), gomock.Eq(args.request.Filters), gomock.Eq(uint64(0)), gomock.Eq(uint64(100))).
					Return(want.Users, nil)
			},
			want: &userServiceV1.ListUsersResponse{Users: []*v1.User{{
				Id:        "abc",
				FirstName: "John",
				LastName:  "Gopher",
				Nickname:  "Goopher",
				Password:  "A-password",
				Email:     "jon@gopher.com",
				Country:   "GBR",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: nil,
			},
				{
					Id:        "abc",
					FirstName: "Beth",
					LastName:  "Gopher",
					Nickname:  "Goopher",
					Password:  "A-password",
					Email:     "beth@gopher.com",
					Country:   "GBR",
					CreatedAt: timestamppb.Now(),
					UpdatedAt: nil,
				},
			}},
		},
		{
			name: "should be able to successfully list users without filters",
			args: args{
				request: &userServiceV1.ListUsersRequest{
					Filters: nil, Offset: 0, Limit: 100},
			},
			setup: func(mockService *mocks.MockService, args args,
				want *userServiceV1.ListUsersResponse) {
				mockService.
					EXPECT().
					ListUsers(gomock.Any(), nil, gomock.Eq(uint64(0)), gomock.Eq(uint64(100))).
					Return(want.Users, nil)
			},
			want: &userServiceV1.ListUsersResponse{Users: []*v1.User{{
				Id:        "abc",
				FirstName: "John",
				LastName:  "Gopher",
				Nickname:  "Goopher",
				Password:  "A-password",
				Email:     "jon@gopher.com",
				Country:   "GBR",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: nil,
			},
				{
					Id:        "abc",
					FirstName: "Beth",
					LastName:  "Gopher",
					Nickname:  "Goopher",
					Password:  "A-password",
					Email:     "beth@gopher.com",
					Country:   "DEU",
					CreatedAt: timestamppb.Now(),
					UpdatedAt: nil,
				},
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args, tt.want)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.ListUsers(context.Background(), tt.args.request)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServer_ListUsers_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.ListUsersRequest
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockService, args args)
		args    args
		wantErr error
	}{
		{
			name: "something went wrong listing in service",
			args: args{
				request: &userServiceV1.ListUsersRequest{
					Filters: &userServiceV1.SelectUserFilters{Countries: []string{"GBR"}}, Offset: 0, Limit: 100},
			},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					ListUsers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("something bad happened"))
			},
			wantErr: status.New(codes.Internal, "oops something went wrong!").Err(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.ListUsers(context.Background(), tt.args.request)
			assert.Nil(t, got)
			require.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestServer_UpdateUser_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.UpdateUserRequest
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockService, args args, want *userServiceV1.UpdateUserResponse)
		args  args
		want  *userServiceV1.UpdateUserResponse
	}{
		{
			name: "should be able to successfully update the user",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup: func(mockService *mocks.MockService, args args,
				want *userServiceV1.UpdateUserResponse) {
				mockService.
					EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(args.request.User), gomock.Eq(args.request.UpdateFields)).
					DoAndReturn(func(ctx context.Context, user *v1.User, field []v1.UpdateUserField) error {
						args.request.User = want.User
						return nil
					})
			},
			want: &userServiceV1.UpdateUserResponse{
				User: &v1.User{
					Id:        "a-user",
					FirstName: "Max",
					LastName:  "Gopher",
					Nickname:  "Goopher",
					Password:  "A-password",
					Email:     "jon@gopher.com",
					Country:   "GBR",
					CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
					UpdatedAt: timestamppb.Now(),
				}},
		}}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args, tt.want)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.UpdateUser(context.Background(), tt.args.request)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServer_UpdateUser_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.UpdateUserRequest
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockService, args args)
		args    args
		wantErr error
	}{
		{
			name:    "should return error when user is not provided",
			args:    args{request: &userServiceV1.UpdateUserRequest{User: nil, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup:   nil,
			wantErr: status.New(codes.InvalidArgument, "user must be provided").Err(),
		},
		{
			name:    "should return error when user id is not provided",
			args:    args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{Id: ""}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup:   nil,
			wantErr: status.New(codes.InvalidArgument, "user id must be provided").Err(),
		},
		{
			name: "should return error when update fields are not provided",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: nil}},
			setup:   nil,
			wantErr: status.New(codes.InvalidArgument, "at least one update field must be provided").Err(),
		},
		{
			name: "should return error when update fields only unspecified",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_UNSPECIFIED}}},
			setup:   nil,
			wantErr: status.New(codes.InvalidArgument, "at least one update field must be provided and not be unspecified value").Err(),
		},
		{
			name: "should return error when update fields are provided twice",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME, v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup:   nil,
			wantErr: status.New(codes.InvalidArgument, "should only input unique update fields").Err(),
		},
		{
			name: "should return error when something went wrong updating",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(args.request.User), gomock.Eq(args.request.UpdateFields)).
					Return(errors.New("something went wrong"))
			},
			wantErr: status.New(codes.Internal, "oops something went wrong!").Err(),
		},
		{
			name: "should return NotFound when service returns NotFound error",
			args: args{request: &userServiceV1.UpdateUserRequest{User: &v1.User{
				Id:        "a-user",
				FirstName: "Max",
				LastName:  "Some field",
				Nickname:  "Nickanme",
				Password:  "a-password",
				Email:     "anemail@gopher.com",
				Country:   "DEU",
				CreatedAt: timestamppb.New(time.Date(2021, 12, 12, 1, 0, 0, 0, time.UTC).UTC()),
				UpdatedAt: nil,
			}, UpdateFields: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME}}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(domain.ErrNoUser)
			},
			wantErr: status.New(codes.NotFound, domain.ErrNoUser.Error()).Err(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.UpdateUser(context.Background(), tt.args.request)
			assert.Nil(t, got)
			require.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestServer_DeleteUser_Success(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.DeleteUserRequest
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockService, args args)
		args  args
		want  *userServiceV1.DeleteUserResponse
	}{
		{
			name: "should be able to successfully get the user",
			args: args{request: &userServiceV1.DeleteUserRequest{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.EXPECT().DeleteUser(gomock.Any(), gomock.Eq(args.request.Id)).Return(nil)
			},
			want: &userServiceV1.DeleteUserResponse{},
		}}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.DeleteUser(context.Background(), tt.args.request)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServer_DeleteUser_Error(t *testing.T) {
	t.Parallel()

	type args struct {
		request *userServiceV1.DeleteUserRequest
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockService, args args)
		args    args
		wantErr error
	}{
		{
			name:    "should throw error if id is missing",
			args:    args{request: &userServiceV1.DeleteUserRequest{Id: ""}},
			setup:   nil,
			wantErr: status.Error(codes.InvalidArgument, "user id must be provided"),
		},
		{
			name:    "should throw error if id is not a uuid",
			args:    args{request: &userServiceV1.DeleteUserRequest{Id: "some-val"}},
			setup:   nil,
			wantErr: status.Error(codes.InvalidArgument, "user id must be in the UUID format"),
		},
		{
			name: "should return NotFound error if service returns domain.ErrNoUser",
			args: args{request: &userServiceV1.DeleteUserRequest{Id: uuid.NewV4().String()}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Return(domain.ErrNoUser)
			},
			wantErr: status.New(codes.NotFound, "user is not found").Err(),
		},
		{
			name: "should return internal error if service returns error",
			args: args{request: &userServiceV1.DeleteUserRequest{Id: uuid.NewV4().String()}},
			setup: func(mockService *mocks.MockService, args args) {
				mockService.
					EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Return(errors.New("some error"))
			},
			wantErr: status.New(codes.Internal, "oops something went wrong!").Err(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockService := mocks.NewMockService(ctrl)
			if tt.setup != nil {
				tt.setup(mockService, tt.args)
			}
			s, err := transportgrpc.NewServer(grpc.NewServer(), mockService)
			require.NoError(t, err)

			got, err := s.DeleteUser(context.Background(), tt.args.request)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
