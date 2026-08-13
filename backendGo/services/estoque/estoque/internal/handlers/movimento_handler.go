package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/services"
)

type MovimentoHandler struct {
	service services.MovimentoService
}

func NewMovimentoHandler(service services.MovimentoService) *MovimentoHandler {
	return &MovimentoHandler{service: service}
}

func (h *MovimentoHandler) Executar(c *gin.Context) {
	var input services.CreateMovimentoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}

	movimento, err := h.service.Executar(input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, movimento)
}
