package domain

import (
	"database/sql"
	"errors"
	"time"

	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	uuid "github.com/kevinburke/go.uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrNoUser = errors.New("user does not exist")

	ErrCreateUserEmailUnique = errors.New("email already exists")
	ErrUserInvalidArgument   = errors.New("invalid request params for modifying/creating user")
)

// User defines a user
type User struct {
	ID        uuid.UUID `db:"id"`
	FirstName string    `db:"first_name" validate:"required"`
	LastName  string    `db:"last_name" validate:"required"`
	Nickname  string    `db:"nickname" validate:"required"`
	Password  string    `db:"password" validate:"required"`
	Email     string    `db:"email" validate:"required,email"`
	// ISO 3166-1 alpha-3
	Country   string       `db:"country" validate:"required,len=3"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

// FromProto converts a proto user into a user.
func (u *User) FromProto(pbUser *v1.User) {
	if pbUser == nil {
		return
	}
	*u = User{
		ID:        uuid.FromStringOrNil(pbUser.Id),
		FirstName: pbUser.FirstName,
		LastName:  pbUser.LastName,
		Nickname:  pbUser.Nickname,
		Email:     pbUser.Email,
		Password:  pbUser.Password,
		Country:   pbUser.Country,
		CreatedAt: pbUser.GetCreatedAt().AsTime(),
	}
	if pbUser.UpdatedAt != nil {
		u.UpdatedAt = sql.NullTime{Time: pbUser.UpdatedAt.AsTime(), Valid: true}
	}
}

// ToProto converts a user into a proto user.
func (u *User) ToProto() *v1.User {
	pbUser := &v1.User{
		Id:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Password:  u.Password,
		Country:   u.Country,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
	if u.UpdatedAt.Valid {
		pbUser.UpdatedAt = timestamppb.New(u.UpdatedAt.Time)
	}
	return pbUser
}
