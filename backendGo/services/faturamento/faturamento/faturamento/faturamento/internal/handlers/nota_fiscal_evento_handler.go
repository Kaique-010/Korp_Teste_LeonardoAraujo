package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/faturamento/internal/services"
)

type NotaFiscalEventoHandler struct {
	notaService  services.NotaFiscalService
	eventService services.NotaFiscalEventoService
}

func NewNotaFiscalEventoHandler(notaService services.NotaFiscalService, eventService services.NotaFiscalEventoService) *NotaFiscalEventoHandler {
	return &NotaFiscalEventoHandler{notaService: notaService, eventService: eventService}
}

func (h *NotaFiscalEventoHandler) Listar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if _, err := h.notaService.Obter(id); err != nil {
		writeError(c, err)
		return
	}

	eventos, err := h.eventService.Listar(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, eventos)
}
