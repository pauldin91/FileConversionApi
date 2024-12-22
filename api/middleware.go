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
			ctx.AbortWithStatusJSON(http.StatusForbidden, "forbidden")
			return
		}
		bearer := strings.Fields(header)

		if len(bearer) != 2 || bearer[0] != authScheme {
			ctx.AbortWithStatusJSON(http.StatusForbidden, "invalid")
			return
		}
		valid, err := server.tokenGenerator.Validate(bearer[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, "server error")
			return
		} else if !valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "server error")
			return
		}

		ctx.Next()
	}
}
