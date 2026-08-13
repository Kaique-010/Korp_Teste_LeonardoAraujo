package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/korp-teste/backendGo/services/faturamento/internal/handlers"
	"github.com/korp-teste/backendGo/services/faturamento/internal/health"
	"github.com/korp-teste/backendGo/services/faturamento/internal/logging"
)

type Handlers struct {
	NotaFiscal *handlers.NotaFiscalHandler
	Evento     *handlers.NotaFiscalEventoHandler
	Cliente    *handlers.ClienteHandler
	Health     *health.Checker
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
			"service": "faturamento",
			"version": logging.Version,
			"links": []gin.H{
				{"rel": "health", "href": "/health", "method": "GET"},
				{"rel": "notas.list", "href": "/notas", "method": "GET"},
				{"rel": "notas.create", "href": "/notas", "method": "POST"},
				{"rel": "clientes.list", "href": "/clientes", "method": "GET"},
				{"rel": "clientes.create", "href": "/clientes", "method": "POST"},
			},
			"rotas": gin.H{
				"clientes": gin.H{
					"listar":  "GET    /clientes",
					"criar":   "POST   /clientes  {nome}",
					"obter":   "GET    /clientes/:id",
					"editar":  "PUT    /clientes/:id  {nome}",
					"excluir": "DELETE /clientes/:id",
				},
				"notas": gin.H{
					"listar":   "GET    /notas",
					"criar":    "POST   /notas  {cliente_id?}",
					"obter":    "GET    /notas/:id",
					"eventos":  "GET    /notas/:id/eventos",
					"add_item": "POST   /notas/:id/itens    {produto_id, descricao?, quantidade, preco_unitario?, preco_vista?, preco_prazo?, desconto?}",
					"rem_item": "DELETE /notas/:id/itens/:item_id",
					"imprimir": "POST   /notas/:id/imprimir   (baixa estoque via RabbitMQ)",
				},
			},
		})
	})

	r.GET("/health", h.Health.Handler)

	clientes := r.Group("/clientes")
	{
		clientes.POST("", h.Cliente.Criar)
		clientes.GET("", h.Cliente.Listar)
		clientes.GET("/:id", h.Cliente.Obter)
		clientes.PUT("/:id", h.Cliente.Atualizar)
		clientes.DELETE("/:id", h.Cliente.Excluir)
	}

	notas := r.Group("/notas")
	{
		notas.POST("", h.NotaFiscal.Criar)
		notas.GET("", h.NotaFiscal.Listar)
		notas.GET("/:id", h.NotaFiscal.Obter)
		notas.GET("/:id/eventos", h.Evento.Listar)
		notas.POST("/:id/itens", h.NotaFiscal.AdicionarItem)
		notas.DELETE("/:id/itens/:item_id", h.NotaFiscal.RemoverItem)
		notas.POST("/:id/imprimir", h.NotaFiscal.Imprimir)
	}
}
