package store

import (
	"context"
	"fmt"
	"github.com/lib/pq"
	"strings"
	"time"

	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/jacktantram/user-service/services/user-service/internal/domain"
	"github.com/jmoiron/sqlx"
	uuid "github.com/kevinburke/go.uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	emailConstraintKey = "users_email_key"
)

func (r Store) GetUser(ctx context.Context, id string) (*v1.User, error) {
	var u domain.User
	if err := r.connFromContext(ctx).QueryRowxContext(ctx, "SELECT * FROM users WHERE id=$1", uuid.FromStringOrNil(id)).StructScan(&u); err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, domain.ErrNoUser
		}
		return nil, err
	}

	return u.ToProto(), nil
}

func (r Store) ListUsers(ctx context.Context, filters *userServiceV1.SelectUserFilters, offset uint64, limit uint64) ([]*v1.User, error) {
	arg := map[string]interface{}{}

	whereBuilder := strings.Builder{}
	if filters != nil {
		whereBuilder.WriteString("WHERE ")
		if len(filters.Countries) != 0 {
			arg["country"] = filters.Countries
			whereBuilder.WriteString("country IN (:country)")
		}
		arg["where"] = whereBuilder.String()
	}
	arg["limit"] = limit
	arg["offset"] = offset

	query, args, err := sqlx.Named(fmt.Sprintf("SELECT * FROM users %s LIMIT :limit OFFSET :offset", whereBuilder.String()), arg)
	if err != nil {
		return nil, err
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return nil, err
	}

	query = r.db.DB.Rebind(query)
	rows, err := r.connFromContext(ctx).Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	users := make([]*v1.User, 0)
	for rows.Next() {
		var user domain.User
		if err = rows.StructScan(&user); err != nil {
			return nil, err
		}

		users = append(users, user.ToProto())
	}
	return users, nil
}

func (r Store) CreateUser(ctx context.Context, user *v1.User) error {

	rows, err := r.connFromContext(ctx).NamedQueryContext(ctx, `
		INSERT INTO users (first_name, last_name, nickname, password, email, country)
		VALUES(:first_name,:last_name,:nickname,:password,:email,:country)
		RETURNING id, created_at;
		`, &domain.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Password:  user.Password,
		Email:     user.Email,
		Country:   user.Country, // todo should really enforce full caps here/ or an enum
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == emailConstraintKey {
				return domain.ErrCreateUserEmailUnique
			}
		}
		return err
	}
	if !rows.Next() {
		return errors.New("row unaffected")
	}
	var (
		id        string
		createdAt time.Time
	)
	if err = rows.Scan(&id, &createdAt); err != nil {
		return errors.Wrap(err, "unable to scan row")
	}
	user.Id = id
	user.CreatedAt = timestamppb.New(createdAt)
	return nil
}

func (r Store) UpdateUser(ctx context.Context, userToUpdate *v1.User, updateFields []v1.UpdateUserField) error {
	if len(updateFields) == 0 {
		return errors.New("missing update fields")
	}

	arg := map[string]interface{}{}

	arg["id"] = userToUpdate.Id

	updateBuilder := strings.Builder{}
	for _, field := range updateFields {
		if field == v1.UpdateUserField_UPDATE_USER_FIELD_FIRST_NAME {
			arg["first_name"] = userToUpdate.FirstName
			updateBuilder.WriteString("first_name=:first_name")
		}
	}

	query, args, err := sqlx.Named(fmt.Sprintf("UPDATE users SET %s, updated_at=now() where id=:id RETURNING updated_at", updateBuilder.String()), arg)
	if err != nil {
		return err
	}

	query = r.db.DB.Rebind(query)

	row := r.connFromContext(ctx).QueryRowxContext(ctx, query, args...)
	if row.Err() != nil {
		return err
	}
	var (
		updatedAt time.Time
	)
	if err = row.Scan(&updatedAt); err != nil {
		return errors.Wrap(err, "unable to scan row")
	}
	userToUpdate.UpdatedAt = timestamppb.New(updatedAt)
	return nil
}

func (r Store) DeleteUser(ctx context.Context, id string) error {
	row, err := r.connFromContext(ctx).ExecContext(ctx, "DELETE FROM users WHERE id=$1", uuid.FromStringOrNil(id))
	if err != nil {
		return err
	}
	affected, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNoUser

	}
	return nil
}
