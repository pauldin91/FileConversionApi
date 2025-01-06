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
		if err := server.storeUploadedFiles(c, files, entryId); err != nil {
			log.Err(err).Msg("Failed to store uploaded files")
		}
		if err := server.createEntryWithDocuments(files, db.CreateEntryWithIdParams{ID: entryId, UserID: user.ID, Operation: operation}); err != nil {
			log.Err(err).Msg("Failed to create entry with documents")
		}
	}()
	c.JSON(http.StatusOK, fmt.Sprintf("Uploaded successfully files %s for %s with id %s", strings.Join(filenames, ","), operation, entryId))
	fmt.Printf("Exit deposit with id %s\n", entryId)
}

func (server *Server) storeUploadedFiles(c *gin.Context, files []*multipart.FileHeader, entryId uuid.UUID) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(files))
	defer close(errorChan)

	for _, file := range files {
		wg.Add(1)
		go func(uploadedFile *multipart.FileHeader) {
			defer wg.Done()
			fullPath := server.storage.GetFilename(entryId.String(), uploadedFile.Filename)
			if err := c.SaveUploadedFile(uploadedFile, fullPath); err != nil {
				log.Err(err).Msgf("unable to upload file %s", uploadedFile.Filename)
				errorChan <- err
			}
		}(file)
	}
	wg.Wait()

	// Check for errors
	if len(errorChan) > 0 {
		return <-errorChan // Return the first error for simplicity
	}
	return nil
}

func (server *Server) createEntryWithDocuments(files []*multipart.FileHeader, entryParams db.CreateEntryWithIdParams) error {
	reqCtx := server.ctx
	entry, err := server.store.CreateEntryWithId(reqCtx, entryParams)
	if err != nil {
		return err
	}

	var params []db.CreateDocumentParams
	for _, file := range files {
		params = append(params, db.CreateDocumentParams{
			EntryID:  entry.ID,
			Filename: filepath.Base(file.Filename),
		})
	}
	batchParams := db.BatchCreateDocumentsParams{
		Column1: make([]uuid.UUID, len(files)),
		Column2: make([]string, len(files)),
	}
	for i := range params {
		batchParams.Column1[i] = params[i].EntryID
		batchParams.Column2[i] = params[i].Filename
	}

	return server.store.BatchCreateDocuments(reqCtx, batchParams)
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
