package main

import (
	"context"
	"github.com/jacktantram/user-service/internal/service"
	"github.com/jacktantram/user-service/internal/store"
	"github.com/jacktantram/user-service/internal/transport/transportgrpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/hellofresh/health-go/v5"
	healthPostgres "github.com/hellofresh/health-go/v5/checks/postgres"
	"github.com/jacktantram/user-service/pkg/driver/v1/config"
	"github.com/jacktantram/user-service/pkg/driver/v1/kafka"
	v1 "github.com/jacktantram/user-service/pkg/driver/v1/postgres"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Cfg represents the services config
type Cfg struct {
	DatabaseURI   string `envconfig:"DATABASE_URI"`
	MigrationPath string `envconfig:"MIGRATION_PATH" default:"/migrations"`

	Kafka struct {
		Hosts []string `envconfig:"KAFKA_HOSTS"`
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// config
	cfg := &Cfg{}

	if err := config.LoadConfig(cfg); err != nil {
		log.WithError(err).Fatalf("unable to load config")
	}
	// database
	client, err := v1.NewClient(cfg.DatabaseURI, "users")
	if err != nil {
		log.WithError(err).Fatal("failed to setup postgres client")
	}
	defer client.DB.Close()

	if err = client.Migrate(cfg.MigrationPath); err != nil {
		log.WithError(err).Fatalf("unable to migrate")
	}
	userStore := store.NewStore(client)

	// kafka
	kafkaProducer, err := kafka.NewSyncProducer(kafka.ProducerConfig{}, cfg.Kafka.Hosts...)
	if err != nil {
		log.WithError(err).Fatal("unable to create kafka producer")
	}

	// grpc
	lis, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{grpc.ChainStreamInterceptor(
		grpcPrometheus.StreamServerInterceptor),
		grpc.ChainUnaryInterceptor(
			grpcPrometheus.UnaryServerInterceptor,
		)}
	/** Turns on recording of handling time
	of RPCs. Histogram metrics can be very expensive for Prometheus
	 to retain and query. todo (look into al **/
	grpcPrometheus.EnableHandlingTimeHistogram(grpcPrometheus.WithHistogramBuckets([]float64{0.1, 0.5, 0.7, 0.9, 0.95, 0.99}))

	grpcServer, err := transportgrpc.NewServer(grpc.NewServer(opts...), service.NewService(userStore, kafkaProducer))
	if err != nil {
		log.WithError(err).Fatal("unable to create new server")
	}
	grpcPrometheus.Register(grpcServer.Server)

	// logging
	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logFormatter.FullTimestamp = true
	log.SetFormatter(logFormatter)

	// http setup

	// add some checks on instance creation
	h, err := health.New(health.WithComponent(health.Component{
		Name:    "user-service",
		Version: "v1.0",
	}), health.WithChecks(
		health.Config{
			Name:      "postgres",
			Timeout:   time.Second * 2,
			SkipOnErr: false,
			Check: healthPostgres.New(healthPostgres.Config{
				DSN: cfg.DatabaseURI,
			}),
		},
	))
	if err != nil {
		log.WithError(err).Fatal("unable to create healthchecks")
	}

	httpServer := http.Server{Addr: ":8080"}
	http.Handle("/health-check/readiness", h.Handler())
	http.HandleFunc("/health-check/liveness", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})
	// Register Prometheus metrics handler.
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info("http server starting on port :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	go func() {
		log.Info("grpc server starting on port :5001")
		if err := grpcServer.Serve(lis); err != nil {
			log.WithError(err).Fatal("unable to serve grpc")
		}
	}()

	log.Print("Server Started")

	<-done
	log.Print("Server Stopping")
	if err = httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	grpcServer.GracefulStop()
	log.Print("Server Shutdown gracefully")

}
