//go:build integration

package integration

import (
	"context"
	"testing"
	"vox/internal/user/voice"
	mod "vox/pkg/models"
	"vox/tests/utils/db"
	"vox/tests/utils/helpers"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSaveNewVoiceReference(t *testing.T) {
	cases := []struct {
		name     string
		u        vars.UserForTests
		userID   string
		text     string
		fileID   string
		path     string
		typeof   string
		nilPool  bool
		nilCtx   bool
		seed     bool
		seedFile bool
		wantErr  bool
	}{
		{
			name:    "valid voice reference",
			u:       vars.User,
			userID:  vars.User.ID,
			text:    "hello world",
			fileID:  "file-123",
			path:    "/voices/file-123.wav",
			typeof:  "wav",
			seed:    true,
			wantErr: false,
		},
		{
			name:     "duplicate file id",
			u:        vars.User,
			userID:   vars.User.ID,
			text:     "hello world",
			fileID:   "file-123",
			path:     "/voices/file-123.wav",
			typeof:   "wav",
			seed:     true,
			seedFile: true,
			wantErr:  true,
		},
		{
			name:    "user does not exist",
			u:       vars.User,
			userID:  "nonexistent-user",
			text:    "hello world",
			fileID:  "file-456",
			path:    "/voices/file-456.wav",
			typeof:  "wav",
			seed:    false,
			wantErr: true,
		},
		{
			name:    "nil pool",
			u:       vars.User,
			userID:  vars.User.ID,
			fileID:  "file-123",
			nilPool: true,
			wantErr: true,
		},
		{
			name:    "nil ctx",
			u:       vars.User,
			userID:  vars.User.ID,
			fileID:  "file-123",
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

			var v voice.PostgresVoice
			if !tc.nilPool {
				v = voice.PostgresVoice{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, vars.User, dbtest)
			}
			if tc.seedFile {
				helpers.InsertFileMetadata(t, tc.fileID, tc.path, tc.typeof, dbtest)
			}

			err := v.SaveNewVoiceReference(ctx, log, tc.userID, tc.text, tc.fileID, tc.path, tc.typeof)
			if tc.wantErr {
				assert.Error(t, err)

				// atomicity check — nothing partial should land on duplicate
				if !tc.nilPool && !tc.nilCtx && !tc.seedFile {
					var count int
					err = dbtest.QueryRow(context.Background(),
						"SELECT COUNT(*) FROM files WHERE id = $1", tc.fileID,
					).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)

					err = dbtest.QueryRow(context.Background(),
						"SELECT COUNT(*) FROM files_and_users WHERE file_id = $1", tc.fileID,
					).Scan(&count)
					assert.NoError(t, err)
					assert.Equal(t, 0, count)
				}
			} else {
				assert.NoError(t, err)

				var fileID string
				err = dbtest.QueryRow(context.Background(),
					"SELECT id FROM files WHERE id = $1", tc.fileID,
				).Scan(&fileID)
				assert.NoError(t, err)
				assert.Equal(t, tc.fileID, fileID)

				var isActive bool
				err = dbtest.QueryRow(context.Background(),
					"SELECT is_active FROM files_and_users WHERE file_id = $1", tc.fileID,
				).Scan(&isActive)
				assert.NoError(t, err)
				assert.True(t, isActive)
			}
		})
	}
}
