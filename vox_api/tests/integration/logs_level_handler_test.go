//go:build integration

package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"vox/tests/utils/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLevelHandler_SetDebug(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":"debug"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, zapcore.DebugLevel, atom.Level())
}

func TestLevelHandler_SetInfo(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":"info"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, zapcore.InfoLevel, atom.Level())
}

func TestLevelHandler_SetWarn(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":"warn"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, zapcore.WarnLevel, atom.Level())
}

func TestLevelHandler_SetError(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":"error"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, zapcore.ErrorLevel, atom.Level())
}

func TestLevelHandler_CaseInsensitive(t *testing.T) {
	cases := []string{"DEBUG", "Info", "WARN", "Error"}
	for _, lvl := range cases {
		t.Run(lvl, func(t *testing.T) {
			api, _ := helpers.NewLogsAPI()
			r := helpers.NewLevelRouter(t, api)

			body := bytes.NewBufferString(`{"level":"` + lvl + `"}`)
			req := httptest.NewRequest(http.MethodPut, "/logs/level", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
		})
	}
}

func TestLevelHandler_InvalidJSON(t *testing.T) {
	api, _ := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`not json at all`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLevelHandler_EmptyBody(t *testing.T) {
	api, _ := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLevelHandler_UnknownLevel(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":"verbose"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, zapcore.InfoLevel, atom.Level())
}

func TestLevelHandler_EmptyLevel(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"level":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, zapcore.InfoLevel, atom.Level())
}

func TestLevelHandler_MissingLevelField(t *testing.T) {
	api, atom := helpers.NewLogsAPI()
	r := helpers.NewLevelRouter(t, api)

	req := httptest.NewRequest(http.MethodPut, "/logs/level", bytes.NewBufferString(`{"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, zapcore.InfoLevel, atom.Level())
}
