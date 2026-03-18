//go:build integration

package integration

import (
	"context"
	"testing"
	"vox/internal/auth"
	mod "vox/pkg/models"
	"vox/tests/utils/db"
	"vox/tests/utils/helpers"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestAddNewManualUser(t *testing.T) {
	cases := []struct {
		name    string
		u       vars.UserForTests
		hash    []byte
		nilPool bool
		nilCtx  bool
		wantErr bool
	}{
		{
			name:    "valid user",
			u:       vars.User,
			hash:    vars.Hash,
			wantErr: false,
		},
		{
			name:    "nil pool",
			u:       vars.User,
			hash:    vars.Hash,
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			u:       vars.User,
			hash:    vars.Hash,
			nilCtx:  true,
			wantErr: true,
		},
		{
			name:    "empty user id",
			u:       vars.UserForTests{Email: "alice@example.com"},
			hash:    vars.Hash,
			wantErr: true,
		},
		{
			name:    "nil hash",
			u:       vars.User,
			hash:    nil,
			wantErr: true,
		},
		{
			name:    "duplicate user id",
			u:       vars.User,
			hash:    vars.Hash,
			wantErr: true,
		},
	}

	db := db.NewTestDB(t)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				_, err := db.Exec(context.Background(), "TRUNCATE users, files CASCADE")
				assert.NoError(t, err)
			})

			log := zaptest.NewLogger(t)

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: db}}
			}

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			if tc.name == "duplicate user id" {
				helpers.InsertAdditionalUserInfo(t, tc.u, db)
			}

			err := dbAuth.AddNewManualUser(ctx, log, auth.UserInfo{
				ID:             tc.u.ID,
				Email:          tc.u.Email,
				Picture:        tc.u.Picture,
				Name:           tc.u.Name,
				UserProviderID: tc.u.UserProviderID,
				ProviderID:     tc.u.ProviderID,
			}, tc.hash)
			if tc.wantErr {
				assert.Error(t, err)

				if !tc.nilPool && !tc.nilCtx && tc.u.ID != "" {
					var count int
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM auth_references WHERE user_id = $1", tc.u.ID).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)

					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", tc.u.ID).Scan(&count)
					assert.NoError(t, err)

					if tc.name == "duplicate user id" {
						assert.Equal(t, 1, count)
					} else {
						assert.Equal(t, 0, count)
					}

					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM users_and_providers WHERE user_id = $1", tc.u.ID).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)
				}
			} else {
				assert.NoError(t, err)

				var id string
				err = db.QueryRow(ctx,
					"SELECT user_id FROM auth_references WHERE user_id = $1", tc.u.ID,
				).Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, tc.u.ID, id)

				err = db.QueryRow(ctx,
					"SELECT id FROM users WHERE id = $1", tc.u.ID,
				).Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, tc.u.ID, id)

				err = db.QueryRow(ctx,
					"SELECT user_id FROM users_and_providers WHERE user_id = $1", tc.u.ID,
				).Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, tc.u.ID, id)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	cases := []struct {
		name           string
		providerID     int
		userProviderID string
		nilPool        bool
		nilCtx         bool
		seed           bool
		wantErr        bool
		wantUser       *auth.UserInfo
	}{
		{
			name:           "valid user",
			providerID:     vars.GOOGLE_PROVIDER_ID,
			userProviderID: "google-123",
			seed:           true,
			wantErr:        false,
			wantUser: &auth.UserInfo{
				ID:             vars.User.ID,
				Email:          vars.User.Email,
				Picture:        vars.User.Picture,
				Name:           vars.User.Name,
				UserProviderID: vars.User.UserProviderID,
				ProviderID:     vars.User.ProviderID,
			},
		},
		{
			name:           "user not found",
			providerID:     vars.GOOGLE_PROVIDER_ID,
			userProviderID: "nonexistent-123",
			seed:           false,
			wantErr:        false,
			wantUser:       &auth.UserInfo{},
		},
		{
			name:           "nil pool",
			providerID:     1,
			userProviderID: "google-123",
			nilPool:        true,
			wantErr:        true,
			wantUser:       nil,
		},
		{
			name:           "nil ctx",
			providerID:     vars.GOOGLE_PROVIDER_ID,
			userProviderID: "google-123",
			nilCtx:         true,
			wantErr:        true,
			wantUser:       nil,
		},
		{
			name:           "zero provider id",
			providerID:     0,
			userProviderID: "google-123",
			wantErr:        true,
			wantUser:       nil,
		},
		{
			name:           "empty user provider id",
			providerID:     vars.GOOGLE_PROVIDER_ID,
			userProviderID: "",
			wantErr:        true,
			wantUser:       nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbtest := db.NewTestDB(t)
			log := zaptest.NewLogger(t)

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
				helpers.InsertProviderUserRef(t, vars.User.ID, tc.providerID, tc.userProviderID, dbtest)
			}

			u, err := dbAuth.GetUser(ctx, log, tc.providerID, tc.userProviderID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, u.ID)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, u.ID)
				assert.Equal(t, tc.wantUser.ID, u.ID)
				assert.Equal(t, tc.wantUser.Email, u.Email)
				assert.Equal(t, tc.wantUser.Name, u.Name)
				assert.Equal(t, tc.wantUser.Picture, u.Picture)
			}
		})
	}
}

func TestAddNewProviderUser(t *testing.T) {
	cases := []struct {
		name    string
		u       vars.UserForTests
		nilPool bool
		nilCtx  bool
		seed    bool
		wantErr bool
	}{
		{
			name:    "valid user",
			u:       vars.User,
			wantErr: false,
		},
		{
			name:    "nil pool",
			u:       vars.User,
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			u:       vars.User,
			nilCtx:  true,
			wantErr: true,
		},
		{
			name:    "duplicate user id",
			u:       vars.User,
			seed:    true,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbtest := db.NewTestDB(t)
			log := zaptest.NewLogger(t)

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: dbtest}}
			}
			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
			}

			err := dbAuth.AddNewProviderUser(ctx, log, auth.UserInfo{
				ID:             tc.u.ID,
				Email:          tc.u.Email,
				Name:           tc.u.Name,
				Picture:        tc.u.Picture,
				ProviderID:     tc.u.ProviderID,
				UserProviderID: tc.u.UserProviderID,
			})
			if tc.wantErr {
				assert.Error(t, err)

				if !tc.nilPool && !tc.nilCtx && tc.u.ID != "" && !tc.seed {
					var count int
					err = dbtest.QueryRow(context.Background(),
						"SELECT COUNT(*) FROM users WHERE id = $1", tc.u.ID,
					).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)

					err = dbtest.QueryRow(context.Background(),
						"SELECT COUNT(*) FROM users_and_providers WHERE user_id = $1", tc.u.ID,
					).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)
				}
			} else {
				assert.NoError(t, err)

				var id string
				err = dbtest.QueryRow(context.Background(),
					"SELECT id FROM users WHERE id = $1", tc.u.ID,
				).Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, tc.u.ID, id)

				err = dbtest.QueryRow(context.Background(),
					"SELECT user_id FROM users_and_providers WHERE user_id = $1", tc.u.ID,
				).Scan(&id)
				assert.NoError(t, err)
				assert.Equal(t, tc.u.ID, id)
			}
		})
	}
}

func TestGetPasswordHash(t *testing.T) {
	cases := []struct {
		name     string
		login    string
		nilPool  bool
		nilCtx   bool
		seed     bool
		wantErr  bool
		wantHash []byte
	}{
		{
			name:     "valid login",
			login:    vars.User.ID,
			seed:     true,
			wantErr:  false,
			wantHash: vars.PasswordHash,
		},
		{
			name:    "user not found",
			login:   "nonexistent-user",
			seed:    false,
			wantErr: true,
		},
		{
			name:    "nil pool",
			login:   vars.User.ID,
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			login:   vars.User.ID,
			nilCtx:  true,
			wantErr: true,
		},
		{
			name:    "empty login",
			login:   "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbtest := db.NewTestDB(t)
			log := zaptest.NewLogger(t)

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
				helpers.InsertPasswordHash(t, vars.User.ID, vars.PasswordHash, dbtest)
			}

			hash, err := dbAuth.GetPasswordHash(ctx, log, tc.login)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantHash, hash)
			}
		})
	}
}

func TestSaveRefreshToken(t *testing.T) {
	cases := []struct {
		name        string
		login       string
		refreshHash string
		nilPool     bool
		nilCtx      bool
		seed        bool
		wantErr     bool
		verifyToken bool
	}{
		{
			name:        "valid save",
			login:       vars.User.ID,
			refreshHash: "newhashvalue",
			seed:        true,
			wantErr:     false,
			verifyToken: true,
		},
		{
			name:        "update existing token",
			login:       vars.User.ID,
			refreshHash: "updatedhashvalue",
			seed:        true,
			wantErr:     false,
			verifyToken: true,
		},
		{
			name:        "user not found — no error but nothing changes",
			login:       "nonexistent-user",
			refreshHash: "sometoken",
			seed:        false,
			wantErr:     true,
			verifyToken: false,
		},
		{
			name:        "nil pool",
			login:       vars.User.ID,
			refreshHash: "sometoken",
			nilPool:     true,
			wantErr:     true,
		},
		{
			name:        "nil ctx",
			login:       vars.User.ID,
			refreshHash: "sometoken",
			nilCtx:      true,
			wantErr:     true,
		},
		{
			name:        "empty login",
			login:       "",
			refreshHash: "sometoken",
			wantErr:     true,
		},
		{
			name:        "empty refresh hash",
			login:       vars.User.ID,
			refreshHash: "",
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbtest := db.NewTestDB(t)
			log := zaptest.NewLogger(t)

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
				helpers.InsertRefreshToken(t, vars.User.ID, vars.RefreshToken, dbtest)
			}

			err := dbAuth.SaveRefreshToken(ctx, log, tc.login, tc.refreshHash)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				if tc.verifyToken {
					var saved string
					err = dbtest.QueryRow(context.Background(),
						"SELECT refresh_token FROM auth WHERE user_id = $1", tc.login,
					).Scan(&saved)
					assert.NoError(t, err)
					assert.Equal(t, tc.refreshHash, saved)
				}
			}
		})
	}
}

func TestGetRefreshToken(t *testing.T) {
	cases := []struct {
		name        string
		login       string
		nilPool     bool
		nilCtx      bool
		seed        bool
		seedRefresh bool
		wantErr     bool
		wantHash    string
	}{
		{
			name:        "valid token retrieved",
			login:       vars.User.ID,
			seed:        true,
			seedRefresh: true,
			wantErr:     false,
			wantHash:    vars.RefreshToken,
		},
		{
			name:    "user not found",
			login:   "nonexistent-user",
			seed:    false,
			wantErr: true,
		},
		{
			name:    "nil pool",
			login:   vars.User.ID,
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			login:   vars.User.ID,
			nilCtx:  true,
			wantErr: true,
		},
		{
			name:    "empty login",
			login:   "",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbtest := db.NewTestDB(t)
			log := zaptest.NewLogger(t)

			ctx := context.Background()
			if tc.nilCtx {
				ctx = nil
			}

			var dbAuth auth.PostgresAuth
			if !tc.nilPool {
				dbAuth = auth.PostgresAuth{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
				helpers.InsertRefreshToken(t, vars.User.ID, vars.RefreshToken, dbtest)
			}

			hash, err := dbAuth.GetRefreshToken(ctx, log, tc.login)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantHash, hash)
			}
		})
	}
}
