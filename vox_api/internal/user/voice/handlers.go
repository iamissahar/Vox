package voice

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"vox/pkg/helpers"
	mod "vox/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type VoiceAPI struct {
	DB  VoiceDB
	Cfg *mod.Config
}

func closeReader(rd io.ReadCloser, log *zap.Logger) {
	err := rd.Close()
	if err != nil {
		log.Error("Failed to close reader", zap.Error(err))
	}
}

// ReferenceHandler godoc
// @Summary      Upload voice reference
// @Description  Receives a raw audio file (webm) as the request body and saves it as a voice reference for the authenticated user along with the provided reference text.
// @Tags         voice
// @Accept       application/octet-stream
// @Param        text_ref  query  string  true  "Reference text associated with the voice sample"
// @Success      200  "Voice reference uploaded successfully"
// @Failure      400  {object}  mod.HttpErrorResponse  "Reference text is missing"
// @Failure      401  {object}  mod.HttpErrorResponse  "Missing or invalid auth cookies (IsAuthorized middleware)"
// @Failure      404  {object}  mod.HttpErrorResponse  "Invalid user ID"
// @Failure      500  {object}  mod.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /user/voice/new [post]
func (v *VoiceAPI) ReferenceHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)

	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	text := ctx.Query("text_ref")
	if text == "" {
		log.Warn("Text ref is invalid", zap.String("text_ref", text))
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	fileID := uuid.New().String()
	path := v.Cfg.StoragePath + "/" + fileID + ".webm"

	f, err := os.Create(path)
	if err != nil {
		log.Error("Failed to create file", zap.String("file_path", path), zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	defer closeReader(f, log)
	defer closeReader(ctx.Request.Body, log)

	rd := bufio.NewReader(ctx.Request.Body)
	wr := bufio.NewWriter(f)

	if _, err = io.Copy(wr, rd); err != nil {
		log.Error("Failed to copy body", zap.Error(err))
		_ = os.Remove(path)
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}
	if err = wr.Flush(); err != nil {
		log.Error("Failed to flush", zap.Error(err))
		_ = os.Remove(path)
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	err = v.DB.SaveNewVoiceReference(ctx.Request.Context(), log, userID, text, fileID, path, "webm")
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	ctx.Status(http.StatusOK)
}
