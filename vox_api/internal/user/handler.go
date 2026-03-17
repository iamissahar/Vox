package user

import (
	"net/http"
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
