package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthOk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := NewChecker("estoque", "1.0.0",
		func(ctx context.Context) error { return nil },
		func() error { return nil },
	)

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)
	g.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	c.Handler(g)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"ok"`)
	assert.Contains(t, w.Body.String(), `"database":"ok"`)
	assert.Contains(t, w.Body.String(), `"rabbitmq":"ok"`)
}

func TestHealthDegradado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := NewChecker("estoque", "1.0.0",
		func(ctx context.Context) error { return errors.New("dial recusado") },
		func() error { return nil },
	)

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)
	g.Request = httptest.NewRequest(http.MethodGet, "/health", nil)
	c.Handler(g)

	require.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"degraded"`)
	assert.Contains(t, w.Body.String(), `"database":"dial recusado"`)
	assert.Contains(t, w.Body.String(), `"rabbitmq":"ok"`)
}
