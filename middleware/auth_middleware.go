package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dutt23/lms/token"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey = "authorization"
	AuthTypeBearer         = "Bearer"
	AuthPayloadKey         = "authotization_payload"
)

func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(AuthorizationHeaderKey)

		if len(authHeader) == 0 {
			err := errors.New("authorization header not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		fields := strings.Fields(authHeader)

		if len(fields) > 2 {
			err := errors.New("invalid auth format supplied")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		authType := fields[0]

		if authType != AuthTypeBearer {
			err := errors.New("auth type not supported by the server")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.Validate(accessToken)

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
		}

		ctx.Set(AuthPayloadKey, payload)
		ctx.Next()
	}
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
