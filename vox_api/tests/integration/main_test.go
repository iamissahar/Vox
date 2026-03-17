//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"vox/tests/utils/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func terminate(container testcontainers.Container, ctx context.Context) {
	err := container.Terminate(ctx)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "postgres",
				"POSTGRES_USER":     "postgres",
				"POSTGRES_DB":       "postgres",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	if err != nil {
		panic(err)
	}
	defer terminate(container, ctx)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	addr := fmt.Sprintf("%s:%s", host, port.Port())

	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://postgres:postgres@%s/postgres?sslmode=disable", addr))
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	db.SetupContainer(m, addr, pool)
	os.Exit(m.Run())
}
