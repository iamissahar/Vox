//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/internal/user/voice"
	"vox/pkg/models"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetMetaReferenceHandler_UserIDNotSet(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetMetaReferenceHandlerRouter(t, api, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference/meta", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMetaReferenceHandler_UserIDWrongType(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetMetaReferenceHandlerRouter(t, api, 12345)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference/meta", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMetaReferenceHandler_GetVoiceReferenceDBError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			return [5]voice.VoiceReference{}, 0, errors.New("db error")
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetMetaReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference/meta", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetMetaReferenceHandler_HappyPath_SingleRecord(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: "/some/path/file.webm"},
			}
			return arr, 0, nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetMetaReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference/meta", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, models.APP_JSON, w.Header().Get("Content-Type"))

	var result []voice.VoiceReference
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "file-abc", result[0].FileID)
	assert.Equal(t, "/some/path/file.webm", result[0].Path)
}

func TestGetMetaReferenceHandler_HappyPath_MultipleRecords(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: "/some/path/file1.webm"},
				{FileID: "file-xyz", Path: "/some/path/file2.webm"},
				{FileID: "file-qwe", Path: "/some/path/file3.webm"},
			}
			return arr, 2, nil // n=2, отдаём arr[:3]
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetMetaReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference/meta", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, models.APP_JSON, w.Header().Get("Content-Type"))

	var result []voice.VoiceReference
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
	require.Len(t, result, 3)
	assert.Equal(t, "file-abc", result[0].FileID)
	assert.Equal(t, "file-xyz", result[1].FileID)
	assert.Equal(t, "file-qwe", result[2].FileID)
}
