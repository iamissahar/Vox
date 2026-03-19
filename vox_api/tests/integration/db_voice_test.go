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
				helpers.InsertFileMetadata(t, tc.fileID, tc.path, tc.typeof, tc.text, dbtest)
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

func TestGetVoiceReference(t *testing.T) {
	cases := []struct {
		name        string
		u           vars.UserForTests
		userID      string
		nilPool     bool
		nilCtx      bool
		seed        bool
		seedFiles   []struct{ fileID, path, typeof, text string }
		wantErr     bool
		wantN       int
		wantResults []voice.VoiceReference
	}{
		{
			name:    "valid user with no files",
			u:       vars.User,
			userID:  vars.User.ID,
			seed:    true,
			wantErr: false,
			wantN:   0,
		},
		{
			name:   "valid user with one file",
			u:      vars.User,
			userID: vars.User.ID,
			seed:   true,
			seedFiles: []struct{ fileID, path, typeof, text string }{
				{"file-001", "/voices/file-001.wav", "wav", "hello world"},
			},
			wantErr: false,
			wantN:   0,
			wantResults: []voice.VoiceReference{
				{FileID: "file-001", Path: "/voices/file-001.wav", Type: "wav", Text: "hello world"},
			},
		},
		{
			name:   "valid user with five files",
			u:      vars.User,
			userID: vars.User.ID,
			seed:   true,
			seedFiles: []struct{ fileID, path, typeof, text string }{
				{"file-001", "/voices/file-001.wav", "wav", "one"},
				{"file-002", "/voices/file-002.wav", "wav", "two"},
				{"file-003", "/voices/file-003.wav", "wav", "three"},
				{"file-004", "/voices/file-004.wav", "wav", "four"},
				{"file-005", "/voices/file-005.wav", "wav", "five"},
			},
			wantErr: false,
			wantN:   4, // last index in a 5-element array
		},
		{
			name:    "nonexistent user returns empty result without error",
			u:       vars.User,
			userID:  "nonexistent-user",
			seed:    false,
			wantErr: false,
			wantN:   0,
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

			var v voice.PostgresVoice
			if !tc.nilPool {
				v = voice.PostgresVoice{Pool: &mod.Pool{Pool: dbtest}}
			}

			if tc.seed {
				helpers.InsertAdditionalUserInfo(t, tc.u, dbtest)
			}

			for _, f := range tc.seedFiles {
				helpers.InsertFileMetadata(t, f.fileID, f.path, f.typeof, f.text, dbtest)
				helpers.InsertFileAndUser(t, tc.userID, f.fileID, dbtest)
			}

			arr, n, err := v.GetVoiceReference(ctx, log, tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantN, n)

			// Verify returned results match seeded files
			if len(tc.wantResults) > 0 {
				for i, want := range tc.wantResults {
					assert.Equal(t, want.FileID, arr[i].FileID)
					assert.Equal(t, want.Path, arr[i].Path)
					assert.Equal(t, want.Type, arr[i].Type)
					assert.Equal(t, want.Text, arr[i].Text)
				}
			}

			// When max files seeded, ensure array is fully populated up to n
			if len(tc.seedFiles) == 5 {
				for i := 0; i <= n; i++ {
					assert.NotEmpty(t, arr[i].FileID)
					assert.NotEmpty(t, arr[i].Path)
				}
			}
		})
	}
}

func TestDeleteVoiceReference(t *testing.T) {
	cases := []struct {
		name     string
		u        vars.UserForTests
		userID   string
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
			name:     "valid delete sets is_active to false",
			u:        vars.User,
			userID:   vars.User.ID,
			fileID:   "file-123",
			path:     "/voices/file-123.wav",
			typeof:   "wav",
			seed:     true,
			seedFile: true,
			wantErr:  false,
		},
		{
			name:    "nonexistent file id is a no op without error",
			u:       vars.User,
			userID:  vars.User.ID,
			fileID:  "file-does-not-exist",
			seed:    true,
			wantErr: false,
		},
		{
			name:     "nonexistent user is a no op without error",
			u:        vars.User,
			userID:   "nonexistent-user",
			fileID:   "file-123",
			path:     "/voices/file-123.wav",
			typeof:   "wav",
			seed:     false,
			seedFile: false,
			wantErr:  false,
		},
		{
			name:     "idempotent deleting already inactive file does not error",
			u:        vars.User,
			userID:   vars.User.ID,
			fileID:   "file-123",
			path:     "/voices/file-123.wav",
			typeof:   "wav",
			seed:     true,
			seedFile: true,
			wantErr:  false,
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
				helpers.InsertAdditionalUserInfo(t, tc.u, dbtest)
			}
			if tc.seedFile {
				helpers.InsertFileMetadata(t, tc.fileID, tc.path, tc.typeof, "", dbtest)
				helpers.InsertFileAndUser(t, tc.userID, tc.fileID, dbtest)
			}

			// For idempotency case: delete once before the actual test call
			if tc.name == "idempotent deleting already inactive file does not error" {
				_ = v.DeleteVoiceReference(ctx, log, tc.userID, tc.fileID)
			}

			err := v.DeleteVoiceReference(ctx, log, tc.userID, tc.fileID)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify is_active was actually flipped when the record exists
			if tc.seedFile && tc.userID == tc.u.ID {
				var isActive bool
				queryErr := dbtest.QueryRow(
					context.Background(),
					"SELECT is_active FROM files_and_users WHERE user_id = $1 AND file_id = $2",
					tc.userID, tc.fileID,
				).Scan(&isActive)
				assert.NoError(t, queryErr)
				assert.False(t, isActive)
			}
		})
	}
}
