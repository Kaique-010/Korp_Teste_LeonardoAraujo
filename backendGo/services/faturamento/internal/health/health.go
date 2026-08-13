package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Checker valida as dependências (banco e RabbitMQ) a cada chamada de /health.
type Checker struct {
	service     string
	version     string
	dbPing      func(ctx context.Context) error
	rabbitAlive func() error
}

func NewChecker(service, version string, dbPing func(ctx context.Context) error, rabbitAlive func() error) *Checker {
	return &Checker{service: service, version: version, dbPing: dbPing, rabbitAlive: rabbitAlive}
}

func (c *Checker) Handler(g *gin.Context) {
	checks := gin.H{}
	saudavel := true

	ctx, cancel := context.WithTimeout(g.Request.Context(), 2*time.Second)
	defer cancel()

	if err := c.dbPing(ctx); err != nil {
		saudavel = false
		checks["database"] = err.Error()
	} else {
		checks["database"] = "ok"
	}

	if err := c.rabbitAlive(); err != nil {
		saudavel = false
		checks["rabbitmq"] = err.Error()
	} else {
		checks["rabbitmq"] = "ok"
	}

	status, statusText := http.StatusOK, "ok"
	if !saudavel {
		status, statusText = http.StatusServiceUnavailable, "degraded"
	}

	g.JSON(status, gin.H{
		"service": c.service,
		"version": c.version,
		"status":  statusText,
		"checks":  checks,
	})
}
