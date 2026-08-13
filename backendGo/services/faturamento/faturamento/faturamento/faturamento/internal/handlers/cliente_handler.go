package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/services"
)

type ClienteHandler struct {
	service services.ClienteService
}

func NewClienteHandler(service services.ClienteService) *ClienteHandler {
	return &ClienteHandler{service: service}
}

func (h *ClienteHandler) Criar(c *gin.Context) {
	var input services.CreateClienteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}
	cliente, err := h.service.Create(input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, cliente)
}

func (h *ClienteHandler) Obter(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	cliente, err := h.service.Get(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, cliente)
}

func (h *ClienteHandler) Listar(c *gin.Context) {
	clientes, err := h.service.List()
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, clientes)
}

func (h *ClienteHandler) Atualizar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input services.UpdateClienteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}
	cliente, err := h.service.Update(id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, cliente)
}

func (h *ClienteHandler) Excluir(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
