package routes

import (
	"github.com/gin-gonic/gin"

	"korp-teste/auth/internal/handlers"
)

func Setup(
	router *gin.Engine,
	usuarioHandler *handlers.UsuarioHandler,
) {
	auth := router.Group("/auth")
	{
		auth.POST("/usuarios", usuarioHandler.Criar)
		auth.POST("/login", usuarioHandler.Login)
	}
}
