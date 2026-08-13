package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
)

func writeError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": apperrors.Internal("Erro interno do servidor"),
	})
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		writeError(c, apperrors.BadRequest("Identificador inválido"))
		return 0, false
	}
	return id, true
}
