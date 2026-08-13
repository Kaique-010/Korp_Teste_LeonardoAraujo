package handlers

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/services"
)

type fakeMovimentoService struct {
	executarFunc func(services.CreateMovimentoInput) (*models.MovimentoEstoque, error)
}

func (f *fakeMovimentoService) Executar(i services.CreateMovimentoInput) (*models.MovimentoEstoque, error) {
	return f.executarFunc(i)
}

func newMovimentoRouter(h *MovimentoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/estoque/movimentos", h.Executar)
	return r
}

func TestHandlerExecutarMovimentoSuccess(t *testing.T) {
	service := &fakeMovimentoService{
		executarFunc: func(i services.CreateMovimentoInput) (*models.MovimentoEstoque, error) {
			return &models.MovimentoEstoque{ID: 1, ProdutoID: i.ProdutoID, Tipo: i.Tipo, Quantidade: i.Quantidade}, nil
		},
	}
	r := newMovimentoRouter(NewMovimentoHandler(service))

	w := performRequest(r, http.MethodPost, "/estoque/movimentos", `{"produto_id":1,"tipo":"ENTRADA","quantidade":5,"origem":"AJUSTE"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandlerExecutarMovimentoJSONInvalido(t *testing.T) {
	service := &fakeMovimentoService{}
	r := newMovimentoRouter(NewMovimentoHandler(service))

	w := performRequest(r, http.MethodPost, "/estoque/movimentos", `{produto_id}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerExecutarMovimentoSaldoInsuficiente(t *testing.T) {
	service := &fakeMovimentoService{
		executarFunc: func(i services.CreateMovimentoInput) (*models.MovimentoEstoque, error) {
			return nil, apperrors.Conflict("Saldo insuficiente")
		},
	}
	r := newMovimentoRouter(NewMovimentoHandler(service))

	w := performRequest(r, http.MethodPost, "/estoque/movimentos", `{"produto_id":1,"tipo":"SAIDA","quantidade":99,"origem":"FATURAMENTO"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandlerExecutarMovimentoProdutoNaoEncontrado(t *testing.T) {
	service := &fakeMovimentoService{
		executarFunc: func(i services.CreateMovimentoInput) (*models.MovimentoEstoque, error) {
			return nil, apperrors.NotFound("Produto não encontrado")
		},
	}
	r := newMovimentoRouter(NewMovimentoHandler(service))

	w := performRequest(r, http.MethodPost, "/estoque/movimentos", `{"produto_id":99,"tipo":"ENTRADA","quantidade":1,"origem":"AJUSTE"}`)

	assert.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "NOT_FOUND")
}
