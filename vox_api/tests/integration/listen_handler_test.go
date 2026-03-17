//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"vox/internal/hub"
	"vox/tests/utils/helpers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestListenHandler_HubMissingFromContext(t *testing.T) {
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouterNoHub(t, api)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.ListenRequest(uuid.New().String()))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListenHandler_WrongTypeInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.InjectLogger(zaptest.NewLogger(t)))
	r.Use(func(ctx *gin.Context) {
		ctx.Set("hub", "not-a-hub-pointer")
		ctx.Next()
	})
	api := &hub.HubAPI{MGR: hub.NewManager()}
	r.GET("/hub/:hub_id/listen", api.ListenHandler)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, helpers.ListenRequest(uuid.New().String()))

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListenHandler_HappyPath_StatusOK(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	// Cancel the request immediately so the stream exits cleanly.
	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()

	w := helpers.NewCloseNotifyRecorder()
	r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListenHandler_HappyPath_AudioMpegContentType(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()

	w := helpers.NewCloseNotifyRecorder()
	r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))

	assert.Equal(t, "audio/mpeg", w.Header().Get("Content-Type"))
}

func TestListenHandler_HappyPath_StreamingHeaders(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()

	w := helpers.NewCloseNotifyRecorder()
	r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))

	assert.Equal(t, "chunked", w.Header().Get("Transfer-Encoding"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestListenHandler_StreamEnds_WhenClientDisconnects(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	reqCtx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		w := helpers.NewCloseNotifyRecorder()
		r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))
	}()

	// Give the stream a moment to start, then disconnect the client.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// handler returned cleanly after client disconnect
	case <-time.After(2 * time.Second):
		t.Fatal("ListenHandler did not return after client disconnect")
	}
}

func TestListenHandler_ConsumerRemovedAfterStreamEnds(t *testing.T) {
	// After the handler returns, the deferred RemoveConsumer must have fired,
	// meaning the hub's consumer map no longer holds the entry (and its Send
	// channel is closed). We verify this indirectly: publishing to the hub
	// after the stream ends must not block or panic.
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	reqCtx, cancel := context.WithCancel(context.Background())
	cancel() // disconnect immediately

	w := helpers.NewCloseNotifyRecorder()
	r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))

	// Must not panic or block after consumer is cleaned up.
	assert.NotPanics(t, func() {
		h.Publish([]byte("post-stream chunk"))
	})
}

func TestListenHandler_PublishedChunksReachClient(t *testing.T) {
	h := hub.NewHub(uuid.New().String())
	defer h.Close()

	api := &hub.HubAPI{MGR: hub.NewManager()}
	r := helpers.NewListenRouter(t, api, h)

	reqCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := helpers.NewCloseNotifyRecorder()
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.ServeHTTP(w, helpers.ListenRequestWithContext(h.ID, reqCtx))
	}()

	// Give the stream goroutine time to register the consumer.
	time.Sleep(20 * time.Millisecond)

	chunk := []byte("audio-data")
	h.Publish(chunk)

	// Give broadcast loop time to deliver.
	time.Sleep(20 * time.Millisecond)
	cancel()

	<-done

	assert.Contains(t, w.Body.String(), "audio-data")
}
