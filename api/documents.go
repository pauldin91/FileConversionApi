package api

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var storageUtil = utils.LocalStorage{}

func (server *Server) convert(c *gin.Context) {
	operation := c.PostForm("operation")
	if strings.ToLower(operation) != "conversion" &&
		strings.ToLower(operation) != "merge" {
		c.JSON(http.StatusBadRequest, errors.New("invalid operation for documents"))
		return
	}
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	token := strings.Fields(c.GetHeader(authHeader))[1]

	claims, err := server.tokenGenerator.Validate(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	user, err := server.store.GetUserByUsername(server.ctx, claims.Username)

	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	entry, err := server.store.CreateEntry(server.ctx, db.CreateEntryParams{UserID: user.ID, Operation: operation})
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	files := form.File["files"]
	filenames := make([]string, len(files))
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		_, err := server.store.CreateDocument(server.ctx, db.CreateDocumentParams{
			EntryID:  entry.ID,
			Filename: filename,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		fullPath := storageUtil.GetFilename(entry.ID.String(), filename)

		if err = c.SaveUploadedFile(file, fullPath); err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Uploaded successfully files %s for %s", strings.Join(filenames, ","), operation))

}

func (server *Server) retrieve(ctx *gin.Context) {
	var req entryRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	parsed, _ := uuid.Parse(req.Id)
	entry, err := server.store.GetEntry(server.ctx, parsed)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	if entry.Status == "processing" {
		ctx.JSON(http.StatusOK, fmt.Sprintf("process of entry with ID %s is being %s", entry.ID.String(), entry.Status))
		return
	} else if entry.Status == "failed" {
		ctx.JSON(http.StatusOK, fmt.Sprintf("process of entry with ID %s failed", entry.ID.String()))
		return
	}

	claims, err := server.tokenGenerator.Validate(strings.Fields(ctx.GetHeader(authHeader))[1])
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, err)
		return
	} else if parsed := entry.UserID.String(); claims.UserId != string(parsed) {
		ctx.JSON(http.StatusForbidden, err)
		return
	}

	filename, err := storageUtil.Retrieve(entry.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	outputName := filepath.Base(filename)
	ctx.FileAttachment(filename, outputName)

}
