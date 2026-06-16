package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rustdesk-api/global"
	"rustdesk-api/http/response"
)

func Limiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		loginLimiter := global.LoginLimiter
		clientIp := c.ClientIP()
		banned, _ := loginLimiter.CheckSecurityStatus(clientIp)
		if banned {
			response.Fail(c, http.StatusLocked, response.TranslateMsg(c, "Banned"))
			c.Abort()
			return
		}
		c.Next()
	}
}
