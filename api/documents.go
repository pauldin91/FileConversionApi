package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	db "github.com/FileConversionApi/db/sqlc"
	"github.com/gin-gonic/gin"
)

func (server *Server) convert(c *gin.Context) {
	operation := c.PostForm("operation")
	if strings.ToLower(operation) != "conversion" && strings.ToLower(operation) != "merge" {
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

	user, err := server.store.GetUserByUsername(context.Background(), claims.Username)

	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	entry, err := server.store.CreateEntry(context.Background(), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	files := form.File["files"]
	filenames := make([]string, len(files))
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		_, err := server.store.CreateDocument(context.Background(), db.CreateDocumentParams{
			EntryID:  entry.ID,
			Filename: filename,
		})
		if err != nil {
			c.JSON(http.StatusNotFound, err.Error())
			return
		}
		if err = c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Uploaded successfully files %s for %s", strings.Join(filenames, ","), operation))

}
