package service

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go -package=mocks

import (
	"context"
	eventsV1 "github.com/jacktantram/user-service/build/go/events/user/v1"
	userServiceV1 "github.com/jacktantram/user-service/build/go/rpc/user/v1"
	v1 "github.com/jacktantram/user-service/build/go/shared/user/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

const (
	// could be moved to a config
	userCreatedTopic = "user-created_v1"
	userUpdatedTopic = "user-updated_v1"
	userDeletedTopic = "user-deleted_v1"

	defaultLimitSize = 100
)

// UserStore CRUD operations for a user.
type UserStore interface {
	GetUser(ctx context.Context, id string) (*v1.User, error)
	ListUsers(ctx context.Context, filters *userServiceV1.SelectUserFilters, offset uint64, limit uint64) ([]*v1.User, error)
	CreateUser(ctx context.Context, user *v1.User) error
	UpdateUser(ctx context.Context, userToUpdate *v1.User, updateFields []v1.UpdateUserField) error
	DeleteUser(ctx context.Context, id string) error
}

// Producer implementation for producing events
type Producer interface {
	ProduceMessage(ctx context.Context, topic string, msg proto.Message) (partition int32, offset int64, err error)
}

// Service defines the service struct.
type Service struct {
	u UserStore
	p Producer
}

// NewService creates a new service
func NewService(store UserStore, p Producer) Service {
	s := &Service{
		u: store,
		p: p,
	}
	return *s
}

// CreateUser attempts to create a new user.
func (s Service) CreateUser(ctx context.Context, user *v1.User) error {
	if err := s.u.CreateUser(ctx, user); err != nil {
		return err
	}

	s.produceMessage(ctx, userCreatedTopic, user.Id, &eventsV1.UserCreatedEvent{User: user})
	return nil
}

// GetUser attempts to fetch a user.
func (s Service) GetUser(ctx context.Context, id string) (*v1.User, error) {
	return s.u.GetUser(ctx, id)
}

// ListUsers attempts to list a set of users.
func (s Service) ListUsers(ctx context.Context, filters *userServiceV1.SelectUserFilters, offset uint64, limit uint64) ([]*v1.User, error) {
	if limit == 0 {
		limit = defaultLimitSize
	}
	return s.u.ListUsers(ctx, filters, offset, limit)
}

// UpdateUser attempts to update a user.
func (s Service) UpdateUser(ctx context.Context, userToUpdate *v1.User, updateFields []v1.UpdateUserField) error {
	if err := s.u.UpdateUser(ctx, userToUpdate, updateFields); err != nil {
		return err
	}

	// don't want to break flow due to publishing error
	s.produceMessage(ctx, userUpdatedTopic, userToUpdate.Id, &eventsV1.UserUpdatedEvent{User: userToUpdate,
		UpdateFields: updateFields})
	return nil
}

// DeleteUser attempts to delete a user.
func (s Service) DeleteUser(ctx context.Context, id string) error {
	u, err := s.u.GetUser(ctx, id)
	if err != nil {
		return errors.Wrap(err, "unable to get user when trying to delete")
	}
	if err = s.u.DeleteUser(ctx, id); err != nil {
		return err
	}

	s.produceMessage(ctx, userDeletedTopic, id, &eventsV1.UserDeletedEvent{User: u})
	return nil

}

// todo - If logger was passed around in context
// userId could potentially be omitted.
func (s Service) produceMessage(ctx context.Context, topicName string, userId string, message proto.Message) {
	_, _, err := s.p.ProduceMessage(ctx, topicName, message)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{
				"user_id": userId, "topic_name": topicName,
			}).Error("unable to produce message")
	}
}
