package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type loginRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) login(ctx *gin.Context) {
	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetUser(context.Background(), req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errors.New("invalid credentials"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errors.New("something gone wrong"))
		return
	}
	err = utils.IsPasswordValid(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errors.New("invalid credentials"))
		return
	}

	token, err := server.tokenGenerator.Generate(user.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errors.New("something gone wrong"))
		return
	}
	rsp := loginResponse{
		AccessToken: token,
	}
	ctx.JSON(http.StatusOK, rsp)

}
