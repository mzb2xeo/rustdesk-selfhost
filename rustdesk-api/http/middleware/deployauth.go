package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"rustdesk-api/global"
	"rustdesk-api/service"
)

// RustOrDeployAuth accepts normal API tokens or short-lived deploy tokens.
func RustOrDeployAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c.GetHeader("Authorization"))
		if token == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		if len(global.Jwt.Key) > 0 {
			uid, _ := service.AllService.UserService.VerifyJWT(token)
			if uid > 0 {
				user := service.AllService.UserService.InfoById(uid)
				if user.Id > 0 && service.AllService.UserService.CheckUserEnable(user) {
					c.Set("curUser", user)
					c.Set("token", token)
					c.Set("authType", "user")
					c.Next()
					return
				}
			}
		}

		user, ut := service.AllService.UserService.InfoByAccessToken(token)
		if user.Id > 0 && service.AllService.UserService.CheckUserEnable(user) {
			c.Set("curUser", user)
			c.Set("token", token)
			c.Set("authType", "user")
			service.AllService.UserService.AutoRefreshAccessToken(ut)
			c.Next()
			return
		}

		dt, err := service.AllService.DeployTokenService.FindValid(token)
		if err == nil && dt != nil {
			user = service.AllService.UserService.InfoById(dt.UserId)
			if user.Id > 0 && service.AllService.UserService.CheckUserEnable(user) {
				c.Set("curUser", user)
				c.Set("token", token)
				c.Set("authType", "deploy")
				c.Set("deployToken", dt)
				c.Next()
				return
			}
		}

		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
	}
}

func extractBearerToken(header string) string {
	if len(header) <= 7 {
		return ""
	}
	return strings.TrimSpace(header[7:])
}
