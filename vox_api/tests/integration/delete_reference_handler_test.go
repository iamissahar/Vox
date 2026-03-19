//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"vox/internal/user/voice"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeleteReferenceHandler_UserIDNotSet(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteReferenceHandler_UserIDWrongType(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, 12345)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteReferenceHandler_FileIDMissing(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteReferenceHandler_GetVoiceReferenceDBError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			return [5]voice.VoiceReference{}, 0, errors.New("db error")
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteReferenceHandler_FileNotOwned(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "other-file-1"},
				{FileID: "other-file-2"},
			}
			return arr, 1, nil // n=1, итерация по arr[:2]
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteReferenceHandler_DeleteVoiceReferenceDBError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc"},
			}
			return arr, 0, nil // n=0, итерация по arr[:1]
		},
		DeleteVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return errors.New("db delete error")
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteReferenceHandler_RemoveFileError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: "/nonexistent/path/file.webm"},
			}
			return arr, 0, nil
		},
		DeleteVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _, _ string) error {
			return nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteReferenceHandler_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "voice-*.webm")
	require.NoError(t, err)
	tmpFile.Close()
	filePath := tmpFile.Name()

	var capturedUserID, capturedFileID string
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: filePath},
			}
			return arr, 0, nil
		},
		DeleteVoiceReferenceF: func(_ context.Context, _ *zap.Logger, userID, fileID string) error {
			capturedUserID = userID
			capturedFileID = fileID
			return nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewDeleteReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "user-123", capturedUserID)
	assert.Equal(t, "file-abc", capturedFileID)
	_, statErr := os.Stat(filePath)
	assert.True(t, os.IsNotExist(statErr), "файл должен быть удалён с диска")
}
