//go:build integration

package integration

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"
	"vox/internal/user/voice"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestReferenceHandler_UserIDNotSet(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewReferenceHandlerRouter(t, api, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference?text_ref=hello", bytes.NewReader([]byte("audio")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReferenceHandler_UserIDWrongType(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewReferenceHandlerRouter(t, api, 12345)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference?text_ref=hello", bytes.NewReader([]byte("audio")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestReferenceHandler_TextRefMissing(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference", bytes.NewReader([]byte("audio")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReferenceHandler_BodyReadError(t *testing.T) {
	cfg := vars.BaseConfig(t.TempDir() + "/")
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: cfg}
	r := helpers.NewReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference?text_ref=hello", iotest.ErrReader(errors.New("read error")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReferenceHandler_SaveNewVoiceReferenceDBError(t *testing.T) {
	db := &mocks.VoiceDB{
		SaveNewVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _, _, _, _, _ string) error {
			return errors.New("db error")
		},
	}
	cfg := vars.BaseConfig(t.TempDir() + "/")
	api := &voice.VoiceAPI{DB: db, Cfg: cfg}
	r := helpers.NewReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference?text_ref=hello", bytes.NewReader([]byte("audio")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestReferenceHandler_HappyPath(t *testing.T) {
	var capturedUserID, capturedText, capturedTypeof string
	db := &mocks.VoiceDB{
		SaveNewVoiceReferenceF: func(_ context.Context, _ *zap.Logger, userID, text, _, _, typeof string) error {
			capturedUserID = userID
			capturedText = text
			capturedTypeof = typeof
			return nil
		},
	}
	cfg := vars.BaseConfig("")
	cfg.StoragePath = t.TempDir() + "/"
	api := &voice.VoiceAPI{DB: db, Cfg: cfg}
	r := helpers.NewReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/voice/reference?text_ref=hello", bytes.NewReader([]byte("audio data")))
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-123", capturedUserID)
	assert.Equal(t, "hello", capturedText)
	assert.Equal(t, "webm", capturedTypeof)
}
