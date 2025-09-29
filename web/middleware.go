package web

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Ctx is an alias you can expand later.
type Ctx = *gin.Context
type Handler = gin.HandlerFunc
type Router = gin.IRouter

// RequestID sets/propagates a request ID.
func RequestID() Handler {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Writer.Header().Set("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

// AccessLog writes a structured access log after the request completes.
func AccessLog(l *slog.Logger) Handler {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		l.Info("http_access",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", dur.Milliseconds(),
			"ip", c.ClientIP(),
			"req_id", c.GetString("request_id"),
		)
	}
}

// RecoveryProblem converts panics to RFC7807 "problem+json".
func RecoveryProblem(l *slog.Logger) Handler {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				l.Error("panic", "error", rec)
				c.Header("Content-Type", "application/problem+json")
				c.JSON(http.StatusInternalServerError, map[string]any{
					"type":   "about:blank",
					"title":  "Internal Server Error",
					"status": http.StatusInternalServerError,
					"detail": "unexpected server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
