package logging

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware loga cada requisição HTTP com método, path, status e duração.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		inicio := time.Now()
		c.Next()
		Info("requisição http", map[string]any{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      c.Writer.Status(),
			"duration_ms": time.Since(inicio).Milliseconds(),
		})
	}
}
