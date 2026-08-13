package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/services"
)

type NotaFiscalHandler struct {
	service services.NotaFiscalService
}

func NewNotaFiscalHandler(service services.NotaFiscalService) *NotaFiscalHandler {
	return &NotaFiscalHandler{service: service}
}

func (h *NotaFiscalHandler) Criar(c *gin.Context) {
	var input services.CriarNotaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		if c.Request.ContentLength > 0 {
			writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
			return
		}
	}

	nota, err := h.service.Criar(input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, nota)
}

func (h *NotaFiscalHandler) Obter(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	nota, err := h.service.Obter(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, nota)
}

func (h *NotaFiscalHandler) Listar(c *gin.Context) {
	notas, err := h.service.Listar()
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, notas)
}

func (h *NotaFiscalHandler) AdicionarItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input services.AdicionarItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}

	nota, err := h.service.AdicionarItem(id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, nota)
}

func (h *NotaFiscalHandler) RemoverItem(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if err != nil {
		writeError(c, apperrors.BadRequest("Identificador de item inválido"))
		return
	}

	if err := h.service.RemoverItem(id, itemID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotaFiscalHandler) Imprimir(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	nota, err := h.service.Imprimir(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, nota)
}
