//go:build integration

package integration

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"vox/internal/user/voice"
	"vox/tests/utils/helpers"
	"vox/tests/utils/mocks"
	"vox/tests/utils/vars"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetFileReferenceHandler_UserIDNotSet(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetFileReferenceHandler_UserIDWrongType(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, 12345)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetFileReferenceHandler_FileIDMissing(t *testing.T) {
	api := &voice.VoiceAPI{DB: &mocks.VoiceDB{}, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetFileReferenceHandler_GetVoiceReferenceDBError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			return [5]voice.VoiceReference{}, 0, errors.New("db error")
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetFileReferenceHandler_FileNotOwned(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "other-file-1"},
				{FileID: "other-file-2"},
			}
			return arr, 1, nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetFileReferenceHandler_OpenFileError(t *testing.T) {
	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: "/nonexistent/path/file.webm"},
			}
			return arr, 0, nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "multipart/form-data")
	body := w.Body.String()
	ct := w.Header().Get("Content-Type")
	boundary := strings.Split(strings.Split(ct, "boundary=")[1], ";")[0]
	assert.NotContains(t, body, "--"+boundary+"--")
}

func TestGetFileReferenceHandler_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()

	file1, err := os.CreateTemp(tmpDir, "voice1-*.webm")
	require.NoError(t, err)
	_, err = file1.Write([]byte("audio data 1"))
	require.NoError(t, err)
	file1.Close()

	file2, err := os.CreateTemp(tmpDir, "voice2-*.webm")
	require.NoError(t, err)
	_, err = file2.Write([]byte("audio data 2"))
	require.NoError(t, err)
	file2.Close()

	db := &mocks.VoiceDB{
		GetVoiceReferenceF: func(_ context.Context, _ *zap.Logger, _ string) ([5]voice.VoiceReference, int, error) {
			arr := [5]voice.VoiceReference{
				{FileID: "file-abc", Path: file1.Name()},
				{FileID: "file-xyz", Path: file2.Name()},
			}
			return arr, 1, nil
		},
	}
	api := &voice.VoiceAPI{DB: db, Cfg: vars.BaseConfig("")}
	r := helpers.NewGetFileReferenceHandlerRouter(t, api, "user-123")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/voice/reference?file_id=file-abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	ct := w.Header().Get("Content-Type")
	assert.Contains(t, ct, "multipart/form-data")

	mediaType, params, err := mime.ParseMediaType(ct)
	require.NoError(t, err)
	assert.Equal(t, "multipart/form-data", mediaType)

	mr := multipart.NewReader(w.Body, params["boundary"])

	part1, err := mr.NextPart()
	require.NoError(t, err)
	body1, err := io.ReadAll(part1)
	require.NoError(t, err)
	assert.Equal(t, "audio data 1", string(body1))
	assert.Equal(t, "audio/webm", part1.Header.Get("Content-Type"))

	part2, err := mr.NextPart()
	require.NoError(t, err)
	body2, err := io.ReadAll(part2)
	require.NoError(t, err)
	assert.Equal(t, "audio data 2", string(body2))
	assert.Equal(t, "audio/webm", part2.Header.Get("Content-Type"))

	_, err = mr.NextPart()
	assert.Equal(t, io.EOF, err)
}
