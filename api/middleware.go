package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authHeader = "authorization"
	authScheme = "bearer"
)

func (server *Server) authorize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader(authHeader)
		if len(header) == 0 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		bearer := strings.Fields(header)

		if len(bearer) != 2 || strings.ToLower(bearer[0]) != authScheme {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "invalid")
			return
		}
		valid, err := server.tokenGenerator.Validate(bearer[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "token invalid error")
			return
		} else if valid == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "invalid")
			return
		}

		ctx.Next()
	}
}
