//go:build integration
// +build integration

package store_test

import (
	"context"
	"fmt"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/jacktantram/user-service/services/user-service/internal/domain"
	uuid "github.com/kevinburke/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

func TestStore_CreateUser(t *testing.T) {
	t.Run("should be able to create a user", func(t *testing.T) {
		var (
			user = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)

		require.NoError(t, testStore.CreateUser(context.Background(), user))

		assert.NotEmpty(t, user.Id)
		assert.NotNil(t, user.CreatedAt)
	})

	t.Run("should throw constraint error for duplicate email on create", func(t *testing.T) {
		var (
			user = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)
		require.NoError(t, testStore.CreateUser(context.Background(), user))

		err := testStore.CreateUser(context.Background(), user)
		require.Error(t, err)
		assert.Error(t, domain.ErrCreateUserEmailUnique, err)
	})
}

func TestStore_GetUser(t *testing.T) {
	t.Run("should successfully get a user", func(t *testing.T) {
		var (
			user = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)

		require.NoError(t, testStore.CreateUser(context.Background(), user))

		u, err := testStore.GetUser(context.Background(), user.GetId())
		require.NoError(t, err)
		// TODO proto equal not working
		assert.Equal(t, user.Id, u.Id)
		assert.Equal(t, user.Email, u.Email)
		assert.Equal(t, user.FirstName, u.FirstName)
		assert.Equal(t, user.LastName, u.LastName)
		assert.Equal(t, user.Nickname, u.Nickname)
		assert.Equal(t, user.Password, u.Password)
		assert.Equal(t, user.Country, u.Country)
		assert.Equal(t, user.CreatedAt, u.CreatedAt)
	})
	t.Run("should return error for unknown user", func(t *testing.T) {
		_, err := testStore.GetUser(context.Background(), uuid.NewV4().String())
		require.Error(t, err)
		assert.Equal(t, domain.ErrNoUser, err)
	})
}

func TestStore_ListUsers(t *testing.T) {
	t.Run("should successfully get a list of users", func(t *testing.T) {
		var (
			user1 = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
			user2 = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)
		require.NoError(t, testStore.CreateUser(context.Background(), user1))
		require.NoError(t, testStore.CreateUser(context.Background(), user2))

		users, err := testStore.ListUsers(context.Background(), nil, 0, 100)
		require.NoError(t, err)
		assert.NotEmpty(t, users)
	})
	t.Run("should successfully get a list of users and filter by country code", func(t *testing.T) {
		var (
			user1 = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "GBK",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)
		require.NoError(t, testStore.CreateUser(context.Background(), user1))

		users, err := testStore.ListUsers(context.Background(), &userServiceV1.SelectUserFilters{Countries: []string{"GBK"}}, 0, 100)
		require.NoError(t, err)
		assert.Len(t, users, 1)

		// TODO proto equal not working
		assert.Equal(t, user1.Id, users[0].Id)
		assert.Equal(t, user1.Email, users[0].Email)
		assert.Equal(t, user1.FirstName, users[0].FirstName)
		assert.Equal(t, user1.LastName, users[0].LastName)
		assert.Equal(t, user1.Nickname, users[0].Nickname)
		assert.Equal(t, user1.Password, users[0].Password)
		assert.Equal(t, user1.Country, users[0].Country)
		assert.Equal(t, user1.CreatedAt, users[0].CreatedAt)
	})

	t.Run("should return no users if users exist", func(t *testing.T) {
		users, err := testStore.ListUsers(context.Background(), &userServiceV1.SelectUserFilters{Countries: []string{"UTC", "TUV"}}, 0, 100)
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestStoreDeleteUser(t *testing.T) {
	t.Run("should successfully delete a user", func(t *testing.T) {
		var (
			user = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: timestamppb.Now(),
			}
		)

		require.NoError(t, testStore.CreateUser(context.Background(), user))

		err := testStore.DeleteUser(context.Background(), user.GetId())
		require.NoError(t, err)

		_, err = testStore.GetUser(context.Background(), user.GetId())
		require.Error(t, err)
		assert.Equal(t, domain.ErrNoUser, err)

	})
	t.Run("should return error for user that does not exist", func(t *testing.T) {
		err := testStore.DeleteUser(context.Background(), uuid.NewV4().String())
		require.Error(t, err)
		assert.Equal(t, domain.ErrNoUser, err)
	})
}

func TestStore_UpdateUser(t *testing.T) {
	t.Parallel()
	t.Run("should successfully update a user", func(t *testing.T) {
		var (
			user = &v1.User{
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: nil,
			}
			newName = "Gophie"
		)
		require.NoError(t, testStore.CreateUser(context.Background(), user))

		user.FirstName = newName

		err := testStore.UpdateUser(context.Background(), user,
			[]v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME})
		require.NoError(t, err)

		u, err := testStore.GetUser(context.Background(), user.GetId())
		require.NoError(t, err)
		assert.Equal(t, newName, u.FirstName)
		assert.NotNil(t, user.UpdatedAt)
	})
	t.Run("should throw an error if user does not exist", func(t *testing.T) {
		var (
			user = &v1.User{
				Id:        uuid.NewV4().String(),
				FirstName: "Sopme",
				LastName:  "asdasd",
				Nickname:  "a-nickname",
				Password:  "a-password",
				Email:     fmt.Sprintf("anemail-%s@.com", uuid.NewV4().String()),
				Country:   "DEU",
				CreatedAt: timestamppb.Now(),
				UpdatedAt: nil,
			}
			newName = "Gophie"
		)

		user.FirstName = newName

		err := testStore.UpdateUser(context.Background(), user,
			[]v1.UpdateUserField{v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME})
		require.Error(t, err)
		assert.Equal(t, domain.ErrNoUser, err)
	})

}
