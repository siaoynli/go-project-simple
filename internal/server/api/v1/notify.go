package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/siaoynli/go-project-simple/internal/pkg/errcode"
	"github.com/siaoynli/go-project-simple/internal/server/api/api_response"
)

func ProductMsgCallback(c *gin.Context) {
	appG := api_response.Gin{C: c}
	appG.ResponseOk(errcode.ErrCodes.ErrNo, nil)
}

func ProductMsgBatchCallback(c *gin.Context) {
	appG := api_response.Gin{C: c}
	appG.ResponseOk(errcode.ErrCodes.ErrNo, nil)
}
