package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/services"
)

type fakeProdutoService struct {
	createFunc func(services.CreateProdutoInput) (*models.Produto, error)
	getFunc    func(uint64) (*models.Produto, error)
	listFunc   func() ([]models.Produto, error)
	updateFunc func(uint64, services.UpdateProdutoInput) (*models.Produto, error)
	deleteFunc func(uint64) error
}

func (f *fakeProdutoService) Create(i services.CreateProdutoInput) (*models.Produto, error) {
	return f.createFunc(i)
}

func (f *fakeProdutoService) Get(id uint64) (*models.Produto, error) {
	return f.getFunc(id)
}

func (f *fakeProdutoService) List() ([]models.Produto, error) {
	return f.listFunc()
}

func (f *fakeProdutoService) Update(id uint64, i services.UpdateProdutoInput) (*models.Produto, error) {
	return f.updateFunc(id, i)
}

func (f *fakeProdutoService) Delete(id uint64) error {
	return f.deleteFunc(id)
}

func newProdutoRouter(h *ProdutoHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/produtos", h.Create)
	r.GET("/produtos", h.List)
	r.GET("/produtos/:id", h.Get)
	r.PUT("/produtos/:id", h.Update)
	r.DELETE("/produtos/:id", h.Delete)
	return r
}

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHandlerCreateProdutoSuccess(t *testing.T) {
	service := &fakeProdutoService{
		createFunc: func(i services.CreateProdutoInput) (*models.Produto, error) {
			return &models.Produto{ID: 1, Codigo: i.Codigo, Descricao: i.Descricao, Saldo: i.Saldo}, nil
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodPost, "/produtos", `{"codigo":"P001","descricao":"Caneta","saldo":5}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, float64(1), body["id"])
}

func TestHandlerCreateProdutoJSONInvalido(t *testing.T) {
	service := &fakeProdutoService{}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodPost, "/produtos", `{codigo}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerCreateProdutoDuplicado(t *testing.T) {
	service := &fakeProdutoService{
		createFunc: func(i services.CreateProdutoInput) (*models.Produto, error) {
			return nil, apperrors.Conflict("Já existe um produto com esse código")
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodPost, "/produtos", `{"codigo":"P001","descricao":"Caneta"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj := body["error"].(map[string]any)
	assert.Equal(t, "CONFLICT", errObj["code"])
}

func TestHandlerGetProdutoSuccess(t *testing.T) {
	service := &fakeProdutoService{
		getFunc: func(id uint64) (*models.Produto, error) {
			return &models.Produto{ID: id, Codigo: "P001", Descricao: "Caneta"}, nil
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodGet, "/produtos/1", "")

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerGetProdutoNotFound(t *testing.T) {
	service := &fakeProdutoService{
		getFunc: func(id uint64) (*models.Produto, error) {
			return nil, apperrors.NotFound("Produto não encontrado")
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodGet, "/produtos/999", "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandlerGetProdutoIDInvalido(t *testing.T) {
	service := &fakeProdutoService{}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodGet, "/produtos/abc", "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandlerUpdateProdutoSuccess(t *testing.T) {
	service := &fakeProdutoService{
		updateFunc: func(id uint64, i services.UpdateProdutoInput) (*models.Produto, error) {
			return &models.Produto{ID: id, Codigo: i.Codigo, Descricao: i.Descricao}, nil
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodPut, "/produtos/1", `{"codigo":"P001","descricao":"Caneta Azul"}`)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandlerDeleteProdutoSuccess(t *testing.T) {
	service := &fakeProdutoService{
		deleteFunc: func(id uint64) error { return nil },
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodDelete, "/produtos/1", "")

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandlerErrorFormatoPadrao(t *testing.T) {
	service := &fakeProdutoService{
		getFunc: func(id uint64) (*models.Produto, error) {
			return nil, apperrors.NotFound("Produto não encontrado")
		},
	}
	r := newProdutoRouter(NewProdutoHandler(service))

	w := performRequest(r, http.MethodGet, "/produtos/999", "")

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	errObj, ok := body["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", errObj["code"])
	assert.Equal(t, "Produto não encontrado", errObj["message"])
}
