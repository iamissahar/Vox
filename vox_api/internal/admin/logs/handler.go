package logs

import (
	"io"
	"net/http"

	mod "vox/pkg/models"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LevelRequest struct {
	Level string `json:"level"`
}

type LogsAPI struct {
	Atomic zap.AtomicLevel
}

func (l *LogsAPI) LevelHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Error("Failed to read request body", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	var lr LevelRequest
	err = sonic.Unmarshal(body, &lr)
	if err != nil {
		log.Warn("Request body is unmarshalable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	if lr.Level == "" {
		log.Warn("Log level is empty")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	var lvl zapcore.Level
	err = lvl.UnmarshalText([]byte(lr.Level))
	if err != nil {
		log.Warn("Invalid log level")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	l.Atomic.SetLevel(lvl)
	ctx.Status(http.StatusNoContent)
}
