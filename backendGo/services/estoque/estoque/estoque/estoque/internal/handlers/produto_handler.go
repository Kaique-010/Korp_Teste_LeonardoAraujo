package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/services"
)

type ProdutoHandler struct {
	service services.ProdutoService
}

func NewProdutoHandler(service services.ProdutoService) *ProdutoHandler {
	return &ProdutoHandler{service: service}
}

func (h *ProdutoHandler) Create(c *gin.Context) {
	var input services.CreateProdutoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}

	produto, err := h.service.Create(input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, produto)
}

func (h *ProdutoHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	produto, err := h.service.Get(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, produto)
}

func (h *ProdutoHandler) List(c *gin.Context) {
	produtos, err := h.service.List()
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, produtos)
}

func (h *ProdutoHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input services.UpdateProdutoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}

	produto, err := h.service.Update(id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, produto)
}

func (h *ProdutoHandler) Delete(c *gin.Context) {
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

func (h *ProdutoHandler) ListarPrecos(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	precos, err := h.service.ListarPrecos(id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, precos)
}

func (h *ProdutoHandler) AtualizarPreco(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input services.AtualizarPrecoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		writeError(c, apperrors.BadRequest("Corpo da requisição inválido"))
		return
	}
	preco, err := h.service.AtualizarPreco(id, input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, preco)
}
