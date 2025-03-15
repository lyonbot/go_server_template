package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	redis_rate "github.com/go-redis/redis_rate/v10"
	"lyonbot.github.com/my_app/infra"
	"lyonbot.github.com/my_app/web"
)

func setupStaticAssets(r *gin.Engine) {
	// 设置静态文件服务
	if os.Getenv("FS_DEV") == "" {
		// 发布模式：使用嵌入的文件系统
		r.StaticFS("/web", http.FS(web.WebFS))
	} else {
		// 开发模式：使用本地文件系统
		workDir, _ := os.Getwd()
		webDir := filepath.Join(workDir, "./web/dist")
		r.Static("/web", webDir)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func IPRateLimitMiddleware() gin.HandlerFunc {
	limter := redis_rate.NewLimiter(infra.Rdb)
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if result, err := limter.Allow(c.Request.Context(), "ipRateLimit:"+ip, redis_rate.PerMinute(60)); err != nil {
			c.AbortWithStatus(500)
			return
		} else {
			c.Header("X-RateLimit-Remaining", strconv.Itoa(int(result.Remaining)))
			c.Header("X-RateLimit-Reset", strconv.Itoa(int(result.ResetAfter.Seconds())))

			if result.Remaining <= 0 {
				c.AbortWithStatus(429)
				return
			}
		}

		c.Next()
	}
}
