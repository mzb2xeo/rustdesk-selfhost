package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"rustdesk-api/global"
)

// Logger log middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		global.Logger.WithFields(
			logrus.Fields{
				"uri":    c.Request.URL,
				"ip":     c.ClientIP(),
				"method": c.Request.Method,
			}).Debug("Request")
		c.Next()
	}
}
