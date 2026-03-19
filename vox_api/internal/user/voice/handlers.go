package voice

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"vox/pkg/helpers"
	mod "vox/pkg/models"

	"github.com/bytedance/sonic"
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

// NewReferenceHandler godoc
// @Summary      Upload voice reference
// @Description  Receives a raw audio file (webm) as the request body and saves it as a voice reference for the authenticated user along with the provided reference text.
// @Tags         voice
// @Accept       application/octet-stream
// @Param        text_ref  query  string  true  "Reference text associated with the voice sample"
// @Success      200  "Voice reference uploaded successfully"
// @Failure      400  {object}  models.HttpErrorResponse  "Reference text is missing"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not authorized"
// @Failure      404  {object}  models.HttpErrorResponse  "Invalid user ID"
// @Failure      500  {object}  models.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /user/voice [post]
func (v *VoiceAPI) NewReferenceHandler(ctx *gin.Context) {
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

// GetMetaReferenceHandler godoc
// @Summary      Get voice reference metadata
// @Description  Returns a list of voice reference metadata records for the specified user
// @Tags         voice
// @Produce      json
// @Success      200  {array}   VoiceReference  "List of voice reference metadata"
//
// @Failure      404  {object}  models.HttpErrorResponse       "Invalid user_id"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not authorized"
// @Failure      500  {object}  models.HttpErrorResponse       "Internal server error"
// @Security     CookieAuth
// @Router       /user/voice/meta [get]
func (v *VoiceAPI) GetMetaReferenceHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)

	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	records, n, err := v.DB.GetVoiceReference(ctx.Request.Context(), log, userID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	metaBytes, err := sonic.Marshal(records[:n+1])
	if err != nil {
		log.Error("Failed to marshal voice reference", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	ctx.Data(http.StatusOK, mod.APP_JSON, metaBytes)
}

// GetFileReferenceHandler godoc
// @Summary      Get voice reference files
// @Description  Returns voice reference audio files as multipart/form-data for the specified user. Verifies that the requested file belongs to the user before streaming.
// @Tags         voice
// @Produce      multipart/form-data
// @Param        file_id  query  string  true  "File identifier to verify ownership"
// @Success      200  {file}    binary          "Multipart response containing audio/webm files"
// @Failure      400  {object}  models.HttpErrorResponse  "Missing file_id"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not authorized"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not the owner of the file"
// @Failure      404  {object}  models.HttpErrorResponse  "Invalid user_id"
// @Failure      500  {object}  models.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /user/voice/file [get]
func (v *VoiceAPI) GetFileReferenceHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)

	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	fileID := ctx.Query("file_id")
	if fileID == "" {
		log.Error("File id is missing")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	records, n, err := v.DB.GetVoiceReference(ctx.Request.Context(), log, userID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	var isOwner bool
	for _, metadata := range records[:n+1] {
		if metadata.FileID == fileID {
			isOwner = true
			break
		}
	}

	if !isOwner {
		log.Error("User is not the owner of the voice reference", zap.String("user_id", userID), zap.String("file_id", fileID))
		ctx.Data(http.StatusForbidden, mod.APP_JSON, mod.HttpError(mod.FORBIDDEN_CODE, mod.FORBIDDEN_MSG))
		return
	}

	mw := multipart.NewWriter(ctx.Writer)
	ctx.Header("Content-Type", mw.FormDataContentType())
	ctx.Status(http.StatusOK)

	for _, metadata := range records[:n+1] {
		file, err := os.Open(metadata.Path)
		if err != nil {
			log.Error("Failed to open voice file", zap.String("path", metadata.Path), zap.Error(err))
			return
		}

		filename := filepath.Base(metadata.Path)
		filePart, err := mw.CreatePart(textproto.MIMEHeader{
			"Content-Disposition": []string{fmt.Sprintf(`form-data; name="files"; filename="%s"`, filename)},
			"Content-Type":        []string{"audio/webm"},
		})
		if err != nil {
			log.Error("Failed to create file part", zap.Error(err))
			if err = file.Close(); err != nil {
				log.Error("Failed to close file", zap.Error(err))
			}
			return
		}

		if _, err = io.Copy(filePart, file); err != nil {
			if err = file.Close(); err != nil {
				log.Error("Failed to close file", zap.Error(err))
			}
			log.Error("Failed to write file part", zap.Error(err))
			return
		}

		if err = file.Close(); err != nil {
			log.Error("Failed to close file", zap.Error(err))
			return
		}
	}

	err = mw.Close()
	if err != nil {
		log.Error("Failed to close multipart writer", zap.Error(err))
	}
}

// DeleteReferenceHandler godoc
// @Summary      Delete voice reference
// @Description  Deletes a voice reference record from the database and removes the associated audio file from storage. Verifies ownership before deletion.
// @Tags         voice
// @Param        file_id  query  string  true  "File identifier to delete"
// @Success      204  "Voice reference successfully deleted"
// @Failure      400  {object}  models.HttpErrorResponse  "Missing file_id"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not authorized"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not the owner of the file"
// @Failure      404  {object}  models.HttpErrorResponse  "Invalid user_id"
// @Failure      500  {object}  models.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /user/voice [delete]
func (v *VoiceAPI) DeleteReferenceHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)

	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusNotFound, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	fileID := ctx.Query("file_id")
	if fileID == "" {
		log.Error("File id is missing")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_URL_CODE, mod.INVALID_URL_MSG))
		return
	}

	ar, n, err := v.DB.GetVoiceReference(ctx.Request.Context(), log, userID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	var isOwner bool
	var i int
	for _, metadata := range ar[:n+1] {
		if metadata.FileID == fileID {
			isOwner = true
			break
		}
		i++
	}

	if !isOwner {
		log.Error("User is not the owner of the voice reference", zap.String("user_id", userID), zap.String("file_id", fileID))
		ctx.Data(http.StatusForbidden, mod.APP_JSON, mod.HttpError(mod.FORBIDDEN_CODE, mod.FORBIDDEN_MSG))
		return
	}

	err = v.DB.DeleteVoiceReference(ctx.Request.Context(), log, userID, fileID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	err = os.Remove(ar[i].Path)
	if err != nil {
		log.Error("Failed to remove voice reference file", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	ctx.Status(http.StatusNoContent)
}
