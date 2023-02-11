package main

import (
	"github.com/jacktantram/user-service/pkg/driver/v1/config"
	"github.com/jacktantram/user-service/pkg/driver/v1/postgres"

	"github.com/jacktantram/user-service/services/user-service/internal/store"
	"github.com/jacktantram/user-service/services/user-service/internal/transport/transportgrpc"

	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Cfg represents the services config
type Cfg struct {
	DatabaseURI   string `envconfig:"DATABASE_URI"`
	MigrationPath string `envconfig:"MIGRATION_PATH" default:"/migrations"`
}

func main() {
	cfg := &Cfg{}

	if err := config.LoadConfig(cfg); err != nil {
		log.WithError(err).Fatalf("unable to load config")
	}

	client, err := postgres.NewClient(cfg.DatabaseURI, "users")
	if err != nil {
		log.WithError(err).Fatal("failed to setup postgres client")
	}
	defer client.DB.Close()

	if err = client.Migrate(cfg.MigrationPath); err != nil {
		log.WithError(err).Fatalf("unable to migrate")
	}

	lis, err := net.Listen("tcp", "localhost:5001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer, err := transportgrpc.NewServer(grpc.NewServer(opts...), store.NewStore(client))
	if err != nil {
		log.WithError(err).Fatal("unable to create new server")
	}

	if err = grpcServer.Serve(lis); err != nil {
		log.WithError(err).Fatal("unable to serve")
	}
}
