package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
)

func TestSolicitarBaixaSuccess(t *testing.T) {
	var received BaixaEstoqueInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/estoque/movimentos", r.URL.Path)
		decodeJSONBody(t, r, &received)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	cliente := NewEstoqueCliente(server.URL)

	err := cliente.SolicitarBaixa(BaixaEstoqueInput{ProdutoID: 10, Tipo: "SAIDA", Quantidade: 2, Origem: "FATURAMENTO"})

	require.NoError(t, err)
	assert.Equal(t, uint64(10), received.ProdutoID)
	assert.Equal(t, "SAIDA", received.Tipo)
	assert.Equal(t, float64(2), received.Quantidade)
	assert.Equal(t, "FATURAMENTO", received.Origem)
}

func TestSolicitarBaixaSaldoInsuficiente(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":{"code":"CONFLICT","message":"Saldo insuficiente para o produto P001"}}`))
	}))
	defer server.Close()

	cliente := NewEstoqueCliente(server.URL)

	err := cliente.SolicitarBaixa(BaixaEstoqueInput{ProdutoID: 10, Quantidade: 99})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
	assert.Equal(t, "Saldo insuficiente para o produto P001", appErr.Message)
}

func TestSolicitarBaixaProdutoNaoEncontrado(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cliente := NewEstoqueCliente(server.URL)

	err := cliente.SolicitarBaixa(BaixaEstoqueInput{ProdutoID: 99, Quantidade: 1})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestSolicitarBaixaServicoIndisponivel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cliente := NewEstoqueCliente(server.URL)

	err := cliente.SolicitarBaixa(BaixaEstoqueInput{ProdutoID: 10, Quantidade: 1})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ServiceUnavailable("").StatusCode, appErr.StatusCode)
}

func TestSolicitarBaixaErroDeRede(t *testing.T) {
	cliente := NewEstoqueCliente("http://127.0.0.1:1")

	err := cliente.SolicitarBaixa(BaixaEstoqueInput{ProdutoID: 10, Quantidade: 1})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ServiceUnavailable("").StatusCode, appErr.StatusCode)
}

func decodeJSONBody(t *testing.T, r *http.Request, out any) {
	t.Helper()
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		t.Fatalf("falha ao decodificar corpo: %v", err)
	}
}
