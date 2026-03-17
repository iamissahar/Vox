package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"vox/db/migrations"
)

var (
	sharedPool *pgxpool.Pool
	sharedAddr string
)

func SetupContainer(m *testing.M, addr string, pool *pgxpool.Pool) {
	sharedAddr = addr
	sharedPool = pool
}

func closeDB(t *testing.T, db *sql.DB) {
	err := db.Close()
	require.NoError(t, err)
}

func NewTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbName := fmt.Sprintf("test_%s", strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_")))

	_, err := sharedPool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := sharedPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s", dbName))
		require.NoError(t, err)
	})

	connStr := fmt.Sprintf("postgres://postgres:postgres@%s/%s?sslmode=disable", sharedAddr, dbName)

	sqlDB, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	defer closeDB(t, sqlDB)

	goose.SetBaseFS(migrations.Files)
	err = goose.SetDialect("postgres")
	require.NoError(t, err)
	err = goose.Up(sqlDB, ".")
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

type UserToAdd struct {
	ID             string
	Email          string
	Name           string
	Picture        string
	ProviderID     int
	UserProviderID string
}

func AddUser(ctx context.Context, t *testing.T, pool *pgxpool.Pool, u UserToAdd, hash []byte) (err error) {
	t.Helper()
	_, err = pool.Exec(ctx, "INSERT INTO users (id, email, name, picture_url) VALUES ($1, $2, $3, $4)", u.ID, u.Email, u.Name, u.Picture)
	if err != nil {
		return
	}
	_, err = pool.Exec(ctx, "INSERT INTO auth_references (user_id, password_hash) VALUES ($1, $2)", u.ID, hash)
	if err != nil {
		return
	}
	_, err = pool.Exec(ctx, "INSERT INTO users_and_providers (user_id, provider_id, user_provider_id) VALUES ($1, $2, $3)", u.ID, u.ProviderID, u.UserProviderID)
	if err != nil {
		return
	}

	return
}

func AddVoice(ctx context.Context, t *testing.T, pool *pgxpool.Pool, userID string, filename string, text string) (err error) {
	t.Helper()
	_, err = pool.Exec(ctx, "INSERT INTO user_voice (user_id, filename, text) VALUES ($1, $2, $3)", userID, filename, text)
	if err != nil {
		return
	}

	return
}
