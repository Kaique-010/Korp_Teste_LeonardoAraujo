package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/services"
)

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

type fakeNotaService struct {
	criarFunc                  func() (*models.NotaFiscal, error)
	obterFunc                  func(uint64) (*models.NotaFiscal, error)
	listarFunc                 func() ([]models.NotaFiscal, error)
	addItemFunc                func(uint64, services.AdicionarItemInput) (*models.NotaFiscal, error)
	removerFunc                func(uint64, uint64) error
	imprimirFunc               func(uint64) (*models.NotaFiscal, error)
	processarResultadoBaixaFunc func(messaging.BaixaResultado) error
}

func (f *fakeNotaService) Criar() (*models.NotaFiscal, error)          { return f.criarFunc() }
func (f *fakeNotaService) Obter(id uint64) (*models.NotaFiscal, error)  { return f.obterFunc(id) }
func (f *fakeNotaService) Listar() ([]models.NotaFiscal, error)         { return f.listarFunc() }
func (f *fakeNotaService) AdicionarItem(id uint64, i services.AdicionarItemInput) (*models.NotaFiscal, error) {
	return f.addItemFunc(id, i)
}
func (f *fakeNotaService) RemoverItem(notaID, itemID uint64) error {
	return f.removerFunc(notaID, itemID)
}
func (f *fakeNotaService) Imprimir(id uint64) (*models.NotaFiscal, error) {
	return f.imprimirFunc(id)
}
func (f *fakeNotaService) ProcessarResultadoBaixa(resultado messaging.BaixaResultado) error {
	if f.processarResultadoBaixaFunc == nil {
		return nil
	}
	return f.processarResultadoBaixaFunc(resultado)
}

func newNotaRouter(h *NotaFiscalHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/notas", h.Criar)
	r.GET("/notas", h.Listar)
	r.GET("/notas/:id", h.Obter)
	r.POST("/notas/:id/itens", h.AdicionarItem)
	r.DELETE("/notas/:id/itens/:item_id", h.RemoverItem)
	r.POST("/notas/:id/imprimir", h.Imprimir)
	return r
}

func TestHandlerCriarNota(t *testing.T) {
	service := &fakeNotaService{
		criarFunc: func() (*models.NotaFiscal, error) {
			return &models.NotaFiscal{ID: 1, Numero: 1, Status: "ABERTA"}, nil
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas", "")

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandlerObterNota(t *testing.T) {
	service := &fakeNotaService{
		obterFunc: func(id uint64) (*models.NotaFiscal, error) {
			return &models.NotaFiscal{ID: id, Numero: 1, Status: "ABERTA"}, nil
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodGet, "/notas/1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerObterNotaNotFound(t *testing.T) {
	service := &fakeNotaService{
		obterFunc: func(id uint64) (*models.NotaFiscal, error) {
			return nil, apperrors.NotFound("Nota fiscal não encontrada")
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodGet, "/notas/99", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandlerAdicionarItem(t *testing.T) {
	service := &fakeNotaService{
		addItemFunc: func(id uint64, i services.AdicionarItemInput) (*models.NotaFiscal, error) {
			return &models.NotaFiscal{ID: id, Numero: 1, Status: "ABERTA",
				Itens: []models.NotaFiscalItem{{ProdutoID: i.ProdutoID, Total: 25}}}, nil
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas/1/itens", `{"produto_id":10,"quantidade":3,"valor_unitario":10,"desconto":5}`)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandlerAdicionarItemJSONInvalido(t *testing.T) {
	service := &fakeNotaService{}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas/1/itens", `{produto_id}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerAdicionarItemNotaFechada(t *testing.T) {
	service := &fakeNotaService{
		addItemFunc: func(id uint64, i services.AdicionarItemInput) (*models.NotaFiscal, error) {
			return nil, apperrors.Conflict("Nota fiscal não está aberta")
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas/1/itens", `{"produto_id":10,"quantidade":1,"valor_unitario":5}`)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandlerRemoverItem(t *testing.T) {
	service := &fakeNotaService{
		removerFunc: func(notaID, itemID uint64) error { return nil },
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodDelete, "/notas/1/itens/5", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandlerRemoverItemIDInvalido(t *testing.T) {
	service := &fakeNotaService{}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodDelete, "/notas/1/itens/abc", "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerImprimir(t *testing.T) {
	service := &fakeNotaService{
		imprimirFunc: func(id uint64) (*models.NotaFiscal, error) {
			return &models.NotaFiscal{ID: id, Numero: 1, Status: "ABERTA"}, nil
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas/1/imprimir", "")

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestHandlerImprimirEstoqueIndisponivel(t *testing.T) {
	service := &fakeNotaService{
		imprimirFunc: func(id uint64) (*models.NotaFiscal, error) {
			return nil, apperrors.ServiceUnavailable("Não foi possível encaminhar a baixa para processamento")
		},
	}
	r := newNotaRouter(NewNotaFiscalHandler(service))

	w := performRequest(r, http.MethodPost, "/notas/1/imprimir", "")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
