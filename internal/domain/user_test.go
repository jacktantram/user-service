package domain

import (
	"database/sql"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	uuid "github.com/kevinburke/go.uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestUser_FromProto(t *testing.T) {
	t.Parallel()
	var (
		id = uuid.FromStringOrNil("a8bdce5a-31dc-4647-98b5-ce9cb343138f")
	)
	t.Run("creating from a nil proto", func(t *testing.T) {
		u := &User{}
		u.FromProto(nil)

		assert.Equal(t, &User{}, u)
	})
	t.Run("creating a user with updated at", func(t *testing.T) {
		pbUser := &v1.User{
			Id:        id.String(),
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
		}
		u := &User{}
		u.FromProto(pbUser)

		assert.Equal(t, &User{
			ID:        id,
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: pbUser.CreatedAt.AsTime(),
			UpdatedAt: sql.NullTime{Valid: true, Time: pbUser.UpdatedAt.AsTime()},
		}, u)
	})
	t.Run("creating a user without updated at", func(t *testing.T) {
		pbUser := &v1.User{
			Id:        id.String(),
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: timestamppb.Now(),
			UpdatedAt: nil,
		}
		u := &User{}
		u.FromProto(pbUser)

		assert.Equal(t, &User{
			ID:        id,
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: pbUser.CreatedAt.AsTime(),
			UpdatedAt: sql.NullTime{Valid: false},
		}, u)
	})
}

func TestUser_ToProto(t *testing.T) {
	t.Parallel()
	var (
		id = uuid.FromStringOrNil("a8bdce5a-31dc-4647-98b5-ce9cb343138f")
	)

	t.Run("creating a user with updated at", func(t *testing.T) {
		u := &User{
			ID:        id,
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}
		pbUser := u.ToProto()

		assert.Equal(t, &v1.User{
			Id:        id.String(),
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: timestamppb.New(u.UpdatedAt.Time),
		}, pbUser)
	})
	t.Run("creating a user without updated at", func(t *testing.T) {
		u := &User{
			ID:        id,
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: time.Now(),
			UpdatedAt: sql.NullTime{Valid: false},
		}
		pbUser := u.ToProto()

		assert.Equal(t, &v1.User{
			Id:        id.String(),
			FirstName: "Sopme",
			LastName:  "asdasd",
			Nickname:  "a-nickname",
			Password:  "a-password",
			Email:     "anemail@gopher.com",
			Country:   "DEU",
			CreatedAt: timestamppb.New(u.CreatedAt),
			UpdatedAt: nil,
		}, pbUser)
	})

}
