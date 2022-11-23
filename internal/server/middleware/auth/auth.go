package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/siaoynli/go-project-simple/config"
	"github.com/siaoynli/go-project-simple/global"
	"github.com/siaoynli/go-project-simple/internal/pkg/errcode"
	"github.com/siaoynli/go-project-simple/internal/pkg/sign"
	"github.com/siaoynli/go-project-simple/internal/repo/mysql/auth_repo"
	"github.com/siaoynli/go-project-simple/internal/server/api/api_response"
	"github.com/siaoynli/go-project-simple/internal/service/auth_service"
	"github.com/siaoynli/pkg/logger"
	"strings"
)

/**
appKey     = "xxx"
secretKey  = "xxx"
encryptParamStr = "param_1=xxx&param_2=xxx&ak="+appKey+"&ts=xxx"

// 自定义验证规则
sn = MD5(secretKey + encryptParamStr + appKey)
*/

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := &api_response.Gin{C: c}
		// header信息校验
		authorization := c.GetHeader(config.HeaderAuthField)
		authorizationDate := c.GetHeader(config.HeaderAuthDateField)
		if len(authorization) == 0 || len(authorizationDate) == 0 {
			appG.ResponseErr(errcode.ErrCodes.ErrAuthenticationHeader)
			c.Abort()
			return
		}
		// 通过签名信息获取 key
		authorizationSplit := strings.Split(authorization, " ")
		if len(authorizationSplit) < 2 {
			appG.ResponseErr(errcode.ErrCodes.ErrAuthenticationHeader)
			c.Abort()
			return
		}

		//解析参数
		err := c.Request.ParseForm()
		if err != nil {
			appG.ResponseErr(errcode.ErrCodes.ErrParams)
			c.Abort()
			return
		}
		key := authorizationSplit[0]
		authService := auth_service.New(global.DB, global.CACHE)
		data, err := authService.DetailByKey(appG, key)
		if err != nil {
			appG.ResponseErr(errcode.ErrCodes.ErrAuthentication)
			c.Abort()
			return
		}

		if data.IsUsed == auth_repo.IsUsedNo {
			appG.ResponseErr(errcode.ErrCodes.ErrAuthentication)
			c.Abort()
			return
		}

		ok, err := sign.New(key, data.Secret, config.HeaderSignTokenTimeoutSeconds).Verify(authorization, authorizationDate,
			c.Request.URL.Path, c.Request.Method, c.Request.Form)
		if err != nil {
			logger.Error("sign verify error")
		}
		if !ok {
			appG.ResponseErr(errcode.ErrCodes.ErrAuthentication)
			c.Abort()
			return
		}
		c.Next()
	}

}
