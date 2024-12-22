package api

import (
	"context"
	"net/http"
	"time"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
)

type userRequest struct {
	Email string `uri:"email" binding:"required,email""`
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {

	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	hashed, err := utils.HashedPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashed,
		FullName:       req.FullName,
		Email:          req.Email,
	}
	user, err := server.store.CreateUser(context.Background(), arg)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	rsp := userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
	}
	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) listUsers(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, "")
}

func (server *Server) getUser(ctx *gin.Context) {
	var req userRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	user, err := server.store.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusNotFound, err)
		return
	}
	ctx.JSON(http.StatusOK, user)
}
