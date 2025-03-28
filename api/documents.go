package api

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func (server *Server) convert(c *gin.Context) {
	reqCtx, cancel := context.WithCancel(server.ctx)
	defer cancel()

	operation := c.PostForm("operation")
	if strings.ToLower(operation) != string(utils.Convert) &&
		strings.ToLower(operation) != string(utils.Merge) {
		c.JSON(http.StatusBadRequest, errors.New("invalid operation for documents"))
		return
	}

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

	user, err := server.store.GetUserByUsername(reqCtx, claims.Username)
	entryId := uuid.New()

	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	files := form.File["files"]
	filenames := make([]string, len(files))
	for i, file := range files {
		filenames[i] = file.Filename
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	go func() {
		server.storeUploadedFiles(c, files, entryId)
		server.createEntryWithDocuments(files, db.CreateEntryWithIdParams{ID: entryId, UserID: user.ID, Operation: operation})
	}()
	c.JSON(http.StatusOK, fmt.Sprintf("Uploaded successfully files %s for %s with id %s", strings.Join(filenames, ","), operation, entryId))

}

func (server *Server) storeUploadedFiles(c *gin.Context, files []*multipart.FileHeader, entryId uuid.UUID) {

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(uploadedFile *multipart.FileHeader) {
			defer wg.Done()
			fullPath := server.storage.GetFilename(entryId.String(), uploadedFile.Filename)
			if err := c.SaveUploadedFile(file, fullPath); err != nil {
				log.Err(err).Msgf("unable to upload file %s", file.Filename)
			}
		}(file)
	}
	wg.Wait()
}

func (server *Server) createEntryWithDocuments(files []*multipart.FileHeader, entryParams db.CreateEntryWithIdParams) {
	reqCtx, cancel := context.WithCancel(server.ctx)
	defer cancel()
	entry, _ := server.store.CreateEntryWithId(reqCtx, entryParams)
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			filename := filepath.Base(file.Filename)
			server.store.CreateDocument(reqCtx, db.CreateDocumentParams{
				EntryID:  entry.ID,
				Filename: filename,
			})
		}()
	}
	wg.Wait()
}

func (server *Server) retrieve(ctx *gin.Context) {
	reqCtx, cancel := context.WithCancel(server.ctx)
	defer cancel()
	var req entryRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	parsed, _ := uuid.Parse(req.Id)
	entry, err := server.store.GetEntry(reqCtx, parsed)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, err)
			return
		}
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	if entry.Status == string(utils.Processing) {
		ctx.JSON(http.StatusOK, fmt.Sprintf("process of entry with ID %s is being %s", entry.ID.String(), entry.Status))
		return
	} else if entry.Status == string(utils.Failed) {
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

	filename, err := server.storage.Retrieve(entry.ID.String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}
	outputName := filepath.Base(filename)
	ctx.FileAttachment(filename, outputName)

}
