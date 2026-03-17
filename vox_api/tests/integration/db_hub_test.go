//go:build integration

package integration

import (
	"context"
	"testing"
	"vox/internal/hub"
	mod "vox/pkg/models"
	"vox/tests/utils/db"
	"vox/tests/utils/helpers"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestGetReference(t *testing.T) {
	cases := []struct {
		name         string
		u            vars.UserForTests
		userID       string
		nilPool      bool
		nilCtx       bool
		seed         bool
		wantErr      bool
		wantFilename string
		wantText     string
	}{
		{
			name:         "valid reference retrieved",
			u:            vars.User,
			userID:       vars.User.ID,
			seed:         true,
			wantErr:      false,
			wantFilename: "voice_ref.wav",
			wantText:     "hello world",
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

			var dbHub hub.PostgresHub
			if !tc.nilPool {
				dbHub = hub.PostgresHub{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
				helpers.InsertVoiceRef(t, vars.User.ID, tc.wantFilename, tc.wantText, dbtest)
			}

			filename, text, err := dbHub.GetReference(ctx, log, tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, filename)
				assert.Empty(t, text)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantFilename, filename)
				assert.Equal(t, tc.wantText, text)
			}
		})
	}
}
