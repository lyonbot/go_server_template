package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"lyonbot.github.com/my_app/infra"
)

// JWT 中间件，可以在路由中使用
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未提供授权信息",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "授权格式错误",
			})
			c.Abort()
			return
		}

		claims, err := infra.ParseJWT(parts[1])
		if claims != nil && claims.Subject == "" {
			err = fmt.Errorf("no subject")
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌: " + err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Set("jwt", parts[1])
		c.Next()
	}
}
