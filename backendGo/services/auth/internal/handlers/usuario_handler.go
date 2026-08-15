package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"korp-teste/auth/internal/jwt"
	"korp-teste/auth/internal/services"
)

type UsuarioHandler struct {
	service    *services.UsuarioService
	jwtService *jwt.JWTService
}

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

func NewUsuarioHandler(
	jwtService *jwt.JWTService,
	service *services.UsuarioService,
) *UsuarioHandler {
	return &UsuarioHandler{
		jwtService: jwtService,
		service:    service,
	}
}

type criarUsuarioRequest struct {
	Nome  string `json:"nome" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

func (h *UsuarioHandler) Criar(c *gin.Context) {
	var req criarUsuarioRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "dados inválidos",
			"details": err.Error(),
		})
		return
	}

	usuario, err := h.service.Criar(
		c.Request.Context(),
		req.Nome,
		req.Email,
		req.Senha,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    usuario.ID,
		"nome":  usuario.Nome,
		"email": usuario.Email,
		"ativo": usuario.Ativo,
	})
}

func (h *UsuarioHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dados inválidos",
		})
		return
	}

	usuario, err := h.service.Autenticar(
		c.Request.Context(),
		req.Email,
		req.Senha,
	)

	if err != nil {
		switch err {
		case services.ErrEmailObrigatorio,
			services.ErrSenhaObrigatoria:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case services.ErrUsuarioInativo,
			services.ErrCredencialInvalida:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "credenciais inválidas",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "erro interno",
			})
		}

		return
	}

	token, err := h.jwtService.Generate(
		usuario.ID,
		usuario.Email,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "erro ao gerar token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"user": gin.H{
			"id":    usuario.ID,
			"nome":  usuario.Nome,
			"email": usuario.Email,
		},
	})
}
