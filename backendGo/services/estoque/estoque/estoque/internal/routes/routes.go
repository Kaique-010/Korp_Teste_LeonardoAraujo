package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/estoque/internal/handlers"
	"github.com/korp-teste/backendGo/services/estoque/internal/health"
	"github.com/korp-teste/backendGo/services/estoque/internal/logging"
)

type Handlers struct {
	Produto   *handlers.ProdutoHandler
	Movimento *handlers.MovimentoHandler
	Health    *health.Checker
}

func charsetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.Next()
	}
}

func Setup(r *gin.Engine, h Handlers) {
	r.Use(logging.Middleware())
	r.Use(charsetMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "estoque",
			"version": logging.Version,
			"links": []gin.H{
				{"rel": "health", "href": "/health", "method": "GET"},
				{"rel": "produtos.list", "href": "/produtos", "method": "GET"},
				{"rel": "produtos.create", "href": "/produtos", "method": "POST"},
				{"rel": "movimentos.executar", "href": "/estoque/movimentos", "method": "POST"},
			},
			"rotas": gin.H{
				"produtos": gin.H{
					"listar":  "GET    /produtos",
					"criar":   "POST   /produtos  {codigo?, descricao, saldo, preco_vista?, preco_prazo?}",
					"obter":   "GET    /produtos/:id",
					"editar":  "PUT    /produtos/:id",
					"excluir": "DELETE /produtos/:id",
					"precos": gin.H{
						"historico": "GET    /produtos/:id/precos",
						"atualizar": "POST   /produtos/:id/precos  {preco_vista, preco_prazo, vigente_em?}",
					},
				},
				"estoque": gin.H{
					"movimento": "POST /estoque/movimentos  {produto_id, tipo:ENTRADA|SAIDA, quantidade, origem, referencia?, idempotency_key?}",
				},
			},
		})
	})

	r.GET("/health", h.Health.Handler)

	produtos := r.Group("/produtos")
	{
		produtos.POST("", h.Produto.Create)
		produtos.GET("", h.Produto.List)
		produtos.GET("/:id", h.Produto.Get)
		produtos.PUT("/:id", h.Produto.Update)
		produtos.DELETE("/:id", h.Produto.Delete)
		produtos.GET("/:id/precos", h.Produto.ListarPrecos)
		produtos.POST("/:id/precos", h.Produto.AtualizarPreco)
	}

	movimentos := r.Group("/estoque/movimentos")
	{
		movimentos.POST("", h.Movimento.Executar)
	}
}
