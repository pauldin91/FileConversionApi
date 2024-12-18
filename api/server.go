package api

import (
	"github.com/FileConversionApi/utils"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    utils.Config
	router *gin.Engine
}

func NewServer(cfg utils.Config,db db.) Server {

}
