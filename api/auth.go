package api

import (
	"errors"
	"net/http"

	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (server *Server) login(ctx *gin.Context) {
	var req loginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	user, err := server.store.GetUserByUsername(server.ctx, req.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
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
	parsed, _ := user.ID.MarshalJSON()

	token, err := server.tokenGenerator.Generate(string(parsed), user.Username, user.Role.String)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errors.New("something gone wrong"))
		return
	}
	rsp := loginResponse{
		AccessToken: token,
	}
	ctx.JSON(http.StatusOK, rsp)

}
