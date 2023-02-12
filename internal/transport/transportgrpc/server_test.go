package transportgrpc_test

import (
	"github.com/golang/mock/gomock"
	"github.com/jacktantram/user-service/internal/transport/transportgrpc"
	"github.com/jacktantram/user-service/internal/transport/transportgrpc/mocks"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"testing"
)

func TestNewServer_Success(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	type args struct {
		server  *grpc.Server
		service transportgrpc.Service
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid server", args{grpc.NewServer(), mocks.NewMockService(ctrl)}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := transportgrpc.NewServer(tt.args.server, tt.args.service)
			assert.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestNewServer_Error(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	type args struct {
		server  *grpc.Server
		service transportgrpc.Service
	}
	tests := []struct {
		name string
		args args
	}{
		{"invalid server missing grpc server", args{nil, mocks.NewMockService(ctrl)}},
		{"invalid server missing mock service", args{grpc.NewServer(), nil}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := transportgrpc.NewServer(tt.args.server, tt.args.service)
			assert.Error(t, err)
			assert.Nil(t, got)
		})
	}
}
