package service_test

import (
	"context"
	"github.com/golang/mock/gomock"
	eventsV1 "github.com/jacktantram/user-service/build/go/events/user/v1"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/jacktantram/user-service/internal/service"
	"github.com/jacktantram/user-service/internal/service/mocks"
	uuid "github.com/kevinburke/go.uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	assert.NotEqual(t, service.Service{}, service.NewService(mocks.NewMockUserStore(ctrl), mocks.NewMockProducer(ctrl)))
}

func TestService_GetUser(t *testing.T) {
	t.Parallel()
	type args struct {
		Id string
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockUserStore, args args, want *v1.User)
		args  args
		want  *v1.User
	}{
		{
			name: "should be able to get a user",
			args: args{
				Id: uuid.FromStringOrNil("a8bdce5a-31dc-4647-98b5-ce9cb343138f").String(),
			},
			setup: func(mockStore *mocks.MockUserStore, args args,
				want *v1.User) {
				mockStore.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq("a8bdce5a-31dc-4647-98b5-ce9cb343138f")).
					Return(want, nil)
			},
			want: &v1.User{
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
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, tt.args, tt.want)
			}
			s := service.NewService(mockUserStore, mockProducer)
			got, err := s.GetUser(context.Background(), tt.args.Id)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	t.Parallel()
	type args struct {
		filters *userServiceV1.SelectUserFilters
		offset  uint64
		limit   uint64
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockUserStore, args args, want []*v1.User)
		args  args
		want  []*v1.User
	}{
		{
			name: "should request a list of users",
			args: args{
				filters: &userServiceV1.SelectUserFilters{Countries: []string{"DBE"}},
				offset:  100,
				limit:   10,
			},
			setup: func(mockStore *mocks.MockUserStore, args args,
				want []*v1.User) {
				mockStore.
					EXPECT().
					ListUsers(gomock.Any(), &userServiceV1.SelectUserFilters{Countries: []string{"DBE"}},
						gomock.Eq(uint64(100)), gomock.Eq(uint64(10))).
					Return(want, nil)
			},
			want: []*v1.User{{
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

		{
			name: "should use default limit if not specified",
			args: args{
				filters: nil,
				offset:  0,
				limit:   0,
			},
			setup: func(mockStore *mocks.MockUserStore, args args,
				want []*v1.User) {
				mockStore.
					EXPECT().
					ListUsers(gomock.Any(), nil, gomock.Eq(uint64(0)), gomock.Eq(uint64(100))).
					Return(want, nil)
			},
			want: []*v1.User{{
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, tt.args, tt.want)
			}
			s := service.NewService(mockUserStore, mockProducer)
			got, err := s.ListUsers(context.Background(), tt.args.filters, tt.args.offset, tt.args.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServiceCreateUser_Success(t *testing.T) {
	t.Parallel()
	type args struct {
		user *v1.User
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args)
		args  args
	}{
		{
			name: "should be able to create a user and publish a created event",
			args: args{
				user: &v1.User{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"},
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				mockStore.
					EXPECT().
					CreateUser(gomock.Any(), args.user).
					Return(nil)
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), "user-created_v1",
						gomock.Eq(&eventsV1.UserCreatedEvent{User: args.user})).
					Return(int32(0), int64(0), nil)
			},
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, mockProducer, tt.args)
			}
			s := service.NewService(mockUserStore, mockProducer)
			require.NoError(t, s.CreateUser(context.Background(), tt.args.user))
		})
	}
}

func TestServiceCreateUser_Error(t *testing.T) {
	t.Parallel()
	type args struct {
		user *v1.User
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args)
		args    args
		wantErr error
	}{
		{
			name: "should return error and not publish message if error creating user",
			args: args{
				user: &v1.User{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"},
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				mockStore.
					EXPECT().
					CreateUser(gomock.Any(), args.user).
					Return(errors.New("some error"))
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			},
			wantErr: errors.New("some error"),
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, mockProducer, tt.args)
			}
			s := service.NewService(mockUserStore, mockProducer)
			err := s.CreateUser(context.Background(), tt.args.user)
			require.Error(t, err)
			assert.Equal(t, tt.wantErr.Error(), err.Error())
		})
	}
}

func TestServiceUpdateUser(t *testing.T) {
	t.Parallel()
	type args struct {
		user           *v1.User
		fieldsToUpdate []v1.UpdateUserField
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args)
		args  args
	}{
		{
			name: "should be able to update a user and publish a updated event",
			args: args{
				user:           &v1.User{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"},
				fieldsToUpdate: []v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME},
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				mockStore.
					EXPECT().
					UpdateUser(gomock.Any(), args.user, args.fieldsToUpdate).
					Return(nil)
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), "user-updated_v1",
						gomock.Eq(&eventsV1.UserUpdatedEvent{User: args.user, UpdateFields: args.fieldsToUpdate})).
					Return(int32(0), int64(0), nil)
			},
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, mockProducer, tt.args)
			}
			s := service.NewService(mockUserStore, mockProducer)
			require.NoError(t, s.UpdateUser(context.Background(), tt.args.user, tt.args.fieldsToUpdate))
		})
	}
}

func TestService_DeleteUser_Success(t *testing.T) {
	t.Parallel()
	type args struct {
		Id string
	}
	tests := []struct {
		name  string
		setup func(mockService *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args)
		args  args
	}{
		{
			name: "should be able to delete a user and publish a deleted event",
			args: args{
				Id: uuid.FromStringOrNil("a8bdce5a-31dc-4647-98b5-ce9cb343138f").String(),
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				existingUser := &v1.User{Id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f"}
				mockStore.
					EXPECT().
					GetUser(gomock.Any(), gomock.Eq("a8bdce5a-31dc-4647-98b5-ce9cb343138f")).
					Return(existingUser, nil)

				mockStore.
					EXPECT().
					DeleteUser(gomock.Any(), gomock.Eq("a8bdce5a-31dc-4647-98b5-ce9cb343138f")).
					Return(nil)
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), "user-deleted_v1",
						gomock.Eq(&eventsV1.UserDeletedEvent{User: existingUser})).
					Return(int32(0), int64(0), nil)
			},
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, mockProducer, tt.args)
			}
			s := service.NewService(mockUserStore, mockProducer)
			require.NoError(t, s.DeleteUser(context.Background(), tt.args.Id))
		})
	}
}

func TestServiceDeleteUser_Error(t *testing.T) {
	t.Parallel()
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		setup   func(mockService *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args)
		args    args
		wantErr error
	}{
		{
			name: "should return error and not delete if unable to fetch user",
			args: args{
				id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f",
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				mockStore.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("get error"))
				mockStore.
					EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Times(0)
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			},
			wantErr: errors.New("get error"),
		},
		{
			name: "should return error and not publish message if error deleting user",
			args: args{
				id: "a8bdce5a-31dc-4647-98b5-ce9cb343138f",
			},
			setup: func(mockStore *mocks.MockUserStore, mockProducer *mocks.MockProducer, args args) {
				mockStore.
					EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(&v1.User{}, nil)
				mockStore.
					EXPECT().
					DeleteUser(gomock.Any(), gomock.Any()).
					Return(errors.New("deleting error"))
				mockProducer.
					EXPECT().
					ProduceMessage(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			},
			wantErr: errors.New("deleting error"),
		}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			mockUserStore := mocks.NewMockUserStore(ctrl)
			mockProducer := mocks.NewMockProducer(ctrl)
			if tt.setup != nil {
				tt.setup(mockUserStore, mockProducer, tt.args)
			}
			s := service.NewService(mockUserStore, mockProducer)
			err := s.DeleteUser(context.Background(), tt.args.id)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr.Error())
		})
	}
}
