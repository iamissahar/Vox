package user

import (
	"io"
	"net/http"
	"strings"
	"vox/internal/hub"
	"vox/pkg/helpers"
	mod "vox/pkg/models"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserAPI struct {
	DB UserDB
}

// InfoHandler godoc
// @Summary      Get user info
// @Description  Returns the profile information of the authenticated user.
// @Tags         user
// @Produce      json
// @Success      200  {object}  UserInfo  "User info returned successfully"
// @Failure      400  {object}  models.HttpErrorResponse  "Invalid user ID"
// @Failure      401  {object}  models.HttpErrorResponse  "Missing or invalid auth cookies (IsAuthorized middleware)"
// @Failure      500  {object}  models.HttpErrorResponse  "Internal server error"
// @Security     CookieAuth
// @Router       /user/info [get]
func (u *UserAPI) InfoHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)
	userID, ok := helpers.IsValString(ctx, "user_id")
	if !ok {
		log.Error("User id is invalid", zap.String("user_id", userID))
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	user, err := u.DB.GetUserInfo(ctx.Request.Context(), log, userID)
	if err != nil {
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	body, err := sonic.Marshal(user)
	if err != nil {
		log.Error("Failed to marshal", zap.Error(err))
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	log.Debug("User info is sent", zap.Any("user_info", user))
	ctx.Data(http.StatusOK, mod.APP_JSON, body)
}

// HubsHandler returns list of hub IDs owned by the user
// @Summary      Get user hubs
// @Description  Returns all hub IDs associated with the authenticated user
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        body  body  object{id=string}  true  "User payload"
// @Success      200  {object}  object{hub_ids=[]string}  "List of hub IDs"
// @Failure      400  {object}  models.HttpErrorResponse  "Invalid request body"
// @Failure      403  {object}  models.HttpErrorResponse  "User is not authenticated"
// @Failure      500  {object}  models.HttpErrorResponse  "Internal server error"
// @Security     BearerAuth
// @Router       /user/hubs [get]
func (uapi *UserAPI) HubsHandler(ctx *gin.Context) {
	log := mod.GetLogger(ctx)

	val, ok := ctx.Get("host_and_hub_cache")
	if !ok {
		log.Error("Invalid host_and_hub_cache type")
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
		return
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Warn("Request body is unreadable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	u := u{}
	err = sonic.Unmarshal(body, &u)
	if err != nil {
		log.Warn("Request body is unmarshalable")
		ctx.Data(http.StatusBadRequest, mod.APP_JSON, mod.HttpError(mod.INVALID_PAYLOAD_CODE, mod.INVALID_PAYLOAD_MSG))
		return
	}

	switch cache := val.(type) {
	case *hub.HostAndHubs:
		ids := cache.GetHubs(u.ID)
		ctx.Data(http.StatusOK, mod.APP_JSON, []byte(`{"hub_ids": ["`+strings.Join(ids, `", "`)+`"]}`))
	default:
		log.Error("Invalid host_and_hub_cache type")
		ctx.Data(http.StatusInternalServerError, mod.APP_JSON, mod.HttpError(mod.INTERNAL_ERROR_CODE, mod.INTERNAL_ERROR_MSG))
	}
}
