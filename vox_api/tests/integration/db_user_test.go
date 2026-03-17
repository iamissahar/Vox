//go:build integration

package integration

import (
	"context"
	"testing"
	"vox/internal/user"
	mod "vox/pkg/models"
	"vox/tests/utils/db"
	"vox/tests/utils/helpers"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestGetUserInfo(t *testing.T) {
	cases := []struct {
		name     string
		u        vars.UserForTests
		userID   string
		nilPool  bool
		nilCtx   bool
		seed     bool
		wantErr  bool
		wantUser *user.UserInfo
	}{
		{
			name:    "valid user info retrieved",
			u:       vars.User,
			userID:  vars.User.ID,
			seed:    true,
			wantErr: false,
			wantUser: &user.UserInfo{
				ID:      vars.User.ID,
				Email:   vars.User.Email,
				Name:    vars.User.Name,
				Picture: vars.User.Picture,
			},
		},
		{
			name:    "user not found",
			u:       vars.User,
			userID:  "nonexistent-user",
			seed:    false,
			wantErr: true,
		},
		{
			name:    "nil pool",
			u:       vars.User,
			userID:  vars.User.ID,
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			u:       vars.User,
			userID:  vars.User.ID,
			nilCtx:  true,
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

			var pu user.PostgresUser
			if !tc.nilPool {
				pu = user.PostgresUser{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
			}

			u, err := pu.GetUserInfo(ctx, log, tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser.ID, u.ID)
				assert.Equal(t, tc.wantUser.Email, u.Email)
				assert.Equal(t, tc.wantUser.Name, u.Name)
				assert.Equal(t, tc.wantUser.Picture, u.Picture)
			}
		})
	}
}
