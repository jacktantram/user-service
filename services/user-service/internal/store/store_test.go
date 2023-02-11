//go:build integration
// +build integration

package store_test

import (
	"github.com/jacktantram/user-service/pkg/driver/v1/postgres"
	"github.com/jacktantram/user-service/services/user-service/internal/store"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var (
	testStore store.Store
)

func TestMain(m *testing.M) {
	postgresClient, err := postgres.NewClient("postgres://postgres:postgres@localhost:5432?sslmode=disable", "users")
	if err != nil {
		log.Fatal(err)
	}
	if err := postgresClient.Migrate("../migrations"); err != nil {
		log.Fatal(err)
	}
	testStore = store.NewStore(postgresClient)
	exitVal := m.Run()
	postgresClient.TruncateTable("users")
	os.Exit(exitVal)

}
