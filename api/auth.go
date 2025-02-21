package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/dutt23/lms/config"
	"github.com/dutt23/lms/middleware"
	service "github.com/dutt23/lms/services"
	"github.com/dutt23/lms/token"
	"github.com/gin-gonic/gin"
)

type authApi struct {
	config *config.AppConfig
  memberService service.MemberService
  tokenMaker token.Maker
}

func NewAuthApi(config *config.AppConfig, memberService service.MemberService, tokenMaker token.Maker) *authApi {
	return &authApi{config, memberService, tokenMaker}
}

type loginUserRequestBody struct {
  Email string `json:"email" binding:"required,email"`
}

type loginUserResponseBody struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expired_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expired_at"`
}

// LoginUser godoc
// @Summary endpoint to login a user
// @Description login a user
// @Tags auth
// @Accept json
// @Produce json
// @Param book body loginUserRequestBody true "Auth data"
// @Success 200 {object} loginUserResponseBody
// @Router /v1/login/user [post]
func (api *authApi) LoginUser(ctx *gin.Context) {
	var req loginUserRequestBody

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := api.memberService.GetMemberByEmail(ctx, req.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	token, payload, err := api.tokenMaker.CreateToken(req.Email, api.config.AccessTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshTokenPayload, err := api.tokenMaker.CreateToken(req.Email, api.config.RefreshTokenDuration)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := loginUserResponseBody{
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenPayload.ExpiredAt,
		AccessToken:           token,
		AccessTokenExpiresAt:  payload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

// CheckAuth godoc
// @Summary endpoint to check auth routes
// @Description auth routes check
// @Tags auth
// @Accept json
// @Produce json
// @Success 200
// @Router /v1/auth/check [post]
func (api *authApi) CheckAuth(ctx *gin.Context) {
  authPayload := ctx.MustGet(middleware.AuthPayloadKey).(*token.Payload)
  fmt.Println(authPayload)
  ctx.Status(http.StatusOK)
}