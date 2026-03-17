package hub

import (
	"io"
	"net/http"
	"os"
	"strings"
	"vox/pkg/helpers"
	mod "vox/pkg/models"

	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	fishaudio "github.com/fishaudio/fish-audio-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type HubAPI struct {
	DB  HubDB
	Cfg *mod.Config
	MGR *Manager
}

func isValHub(ctx *gin.Context, key string) (hub *Hub, ok bool) {
	val, _ok := ctx.Get(key)
	if !_ok {
		return hub, ok
	}

	switch v := val.(type) {
	case *Hub:
		hub = v
		ok = true
	default:
		return hub, ok
	}
	return hub, ok
}

func (h *HubAPI) IsContentTypeValid(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	contentType := ctx.GetHeader("Content-Type")
	if !strings.HasPrefix(contentType, "audio/") && contentType != "application/octet-stream" {
		ctx.Abort()
		log.Warn("Invalid content-type", zap.String("content_type", contentType))
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_CONTENT_TYPE_CODE, mod.INVALID_CONTENT_TYPE_MSG))
		return
	}
	log.Debug("Content-type is valid", zap.String("content_type", contentType))
	ctx.Next()
}

func (h *HubAPI) IsHubIDValid(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	hubID := ctx.Param("hub_id")
	if hubID == "" {
		ctx.Abort()
		log.Error("Invalid hub id", zap.Bool("hub_id_is_empty", hubID == ""))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	hub, ok := h.MGR.Get(hubID)
	if !ok {
		ctx.Abort()
		log.Warn("Invalid hub id", zap.String("hub_id", hubID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}
	ctx.Set("hub", hub)
	log.Debug("Hub id is valid", zap.String("hub_id", hubID))
	ctx.Next()
}

// NewHubHandler godoc
// @Summary      Create a new hub
// @Description  Creates a new hub with the given hub ID. Returns a conflict error if a hub with the same ID already exists.
// @Tags         hub
// @Produce      json
// @Param        hub_id  path  string  true  "Hub ID"
// @Success      201  {object}  map[string]string  "Hub created successfully"
// @Failure      404  {object}  mod.HttpErrorResponse  "Hub ID is empty or invalid"
// @Failure      409  {object}  mod.HttpErrorResponse  "Hub with the given ID already exists"
// @Failure      401  {object}  mod.HttpErrorResponse  "Missing or invalid auth cookies (IsAuthorized middleware)"
// @Failure      415  {object}  mod.HttpErrorResponse  "Invalid content type (IsContentTypeValid middleware)"
// @Security     CookieAuth
// @Router       /hub/{hub_id}/new [post]
func (h *HubAPI) NewHubHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	hubID := ctx.Param("hub_id")
	if hubID == "" {
		log.Error("Invalid hub id", zap.Bool("hub_id_is_empty", hubID == ""))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	_, ok := h.MGR.Get(hubID)
	if ok {
		log.Warn("Hub already exists", zap.String("hub_id", hubID))
		ctx.Data(http.StatusConflict, mod.APP_JSON, mod.HttpError(mod.ENTITY_EXISTS_CODE, mod.ENTITY_EXISTS_MSG))
	} else {
		hubID = h.MGR.New()
		log.Debug("New hub created", zap.String("hub_id", hubID))
		ctx.Data(http.StatusCreated, mod.APP_JSON, []byte(`{"hub_id": "`+hubID+`"}`))
	}
}

// ListenHandler godoc
// @Summary      Listen to audio stream
// @Description  Subscribes the client to a hub's audio stream and delivers synthesized audio chunks in real-time via chunked transfer encoding. The stream ends when the client disconnects or the hub closes.
// @Tags         hub
// @Produce      audio/mpeg
// @Param        hub_id  path  string  true  "Hub ID"
// @Success      200  "Audio stream delivered as chunked audio/mpeg"
// @Failure      404  {object}  mod.HttpErrorResponse  "Invalid hub ID (IsHubIDValid middleware)"
// @Router       /hub/{hub_id}/listen [get]
func (h *HubAPI) ListenHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	hub, ok := isValHub(ctx, "hub")
	if !ok {
		log.Error("Invalid hub id", zap.Any("hub", hub))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	consumerID := uuid.New().String()
	consumer := &Consumer{
		ID:   consumerID,
		Send: make(chan []byte, 128),
	}

	hub.AddConsumer(consumer)
	defer hub.RemoveConsumer(consumerID)

	ctx.Header("Content-Type", "audio/mpeg")
	ctx.Header("Transfer-Encoding", "chunked")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Status(http.StatusOK)

	clientGone := ctx.Request.Context().Done()

	log.Debug("Audio stream started", zap.String("hub_id", hub.ID), zap.String("consumer_id", consumerID))
	ctx.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-consumer.Send:
			if !ok {
				return false
			}
			_, err := w.Write(chunk)
			return err == nil

		case <-clientGone:
			return false
		}
	})
	log.Debug("Audio stream ended", zap.String("hub_id", hub.ID), zap.String("consumer_id", consumerID))
}

func closeReader(rd io.ReadCloser, log *zap.Logger) {
	err := rd.Close()
	if err != nil {
		log.Error("Failed to close reader", zap.Error(err))
	}
}

func (h *HubAPI) FishSDK(ctx *gin.Context) {
	ctx.Set("fish_builder", &BuildHolder{
		client: fishaudio.NewClient(
			fishaudio.WithAPIKey(h.Cfg.FishAudioAPIKey),
			fishaudio.WithBaseURL(h.Cfg.FishAudioBaseURL),
		).TTS,
	})
}

// PublishHandler godoc
// @Summary      Publish audio stream
// @Description  Receives an audio stream, transcribes it via Deepgram, processes the transcription via Groq, and synthesizes speech via Fish Audio. All three operations run concurrently. Requires a valid user and hub in context.
// @Tags         hub
// @Accept       application/octet-stream
// @Param        hub_id  path  string  true  "Hub ID"
// @Param        lang    path  string  true  "Transcription language code (e.g. en, ru)"
// @Success      200  "Audio stream processed successfully"
// @Failure      401  {object}  mod.HttpErrorResponse  "Missing or invalid auth cookies (IsAuthorized middleware)"
// @Failure      404  {object}  mod.HttpErrorResponse  "Invalid user ID or hub"
// @Failure      415  {object}  mod.HttpErrorResponse  "Invalid content type (IsContentTypeValid middleware)"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /hub/{hub_id}/publish [post]
func (h *HubAPI) PublishHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	transcription := NewStringChanBuf(1)
	tokens := NewStringChanBuf(1)
	groqErrors := NewErrorChanBuf(1)
	deepgramErrors := NewErrorChanBuf(1)

	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	hub, ok := isValHub(ctx, "hub")
	if !ok {
		log.Error("Hub is invalid", zap.Any("hub", hub))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	filename, text, err := h.DB.GetReference(ctx.Request.Context(), log, userID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	body, err := os.ReadFile(filename)
	if err != nil {
		log.Error("Failed to read file", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	var fish FishAudio
	if fish_sdk, ok := ctx.Get("fish_builder"); !ok {
		log.Error("fish_builder not found in context")
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	} else {
		switch fb := fish_sdk.(type) {
		case FishBuilder:
			fb.SetReference(body, text)
			fb.SetHub(hub)
			fb.SetTokens(tokens)
			fb.SetLogger(log)
			fish = fb.Get()
		default:
			log.Error("fish_builder is not a FishBuilder")
			ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
			return
		}
	}

	defer closeReader(ctx.Request.Body, log)
	defer h.MGR.Delete(hub.ID)

	g, gctx := errgroup.WithContext(ctx.Request.Context())

	deepgram := Deepgram{
		ApiKey:  h.Cfg.DeepgramAPIKey,
		BaseURL: h.Cfg.DeepgramBaseURL,
		Options: interfaces.LiveTranscriptionOptions{
			Model:       h.Cfg.DeepgramModel,
			Language:    ctx.Param("lang"),
			Channels:    1,
			Endpointing: "true",
			Numerals:    true,
			Punctuate:   true,
		},
		transcription: transcription,
		errors:        deepgramErrors,
		log:           log,
		ctx:           gctx,
	}

	groq := Groq{
		ApiKey:        h.Cfg.GroqAPIKey,
		Model:         h.Cfg.GroqModel,
		BaseURL:       h.Cfg.GroqBaseURL,
		transcription: transcription,
		errors:        groqErrors,
		tokens:        tokens,
		log:           log,
	}
	log.Debug("Audio publishing started", zap.String("user_id", userID), zap.String("hub_id", hub.ID))

	g.Go(func() error { return deepgram.do(ctx.Request.Body) })
	g.Go(func() error { return groq.do(gctx) })
	g.Go(func() error { return fish.Do(gctx) })

	if err := g.Wait(); err != nil {
		transcription.Close()
		deepgramErrors.Close()
		groqErrors.Close()
		tokens.Close()
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	log.Debug("Audio publishing ended", zap.String("user_id", userID), zap.String("hub_id", hub.ID))
	ctx.Status(http.StatusOK)
}
