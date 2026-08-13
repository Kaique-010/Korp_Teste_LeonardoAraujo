package repositories

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/models"
)

func TestMovimentoRepositoryInsertAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovimentoRepository(db)

	movimento := &models.MovimentoEstoque{
		ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 5, Origem: "AJUSTE",
	}
	require.NoError(t, repo.Insert(movimento))
	assert.NotZero(t, movimento.ID)

	found, err := repo.FindByID(movimento.ID)
	require.NoError(t, err)
	assert.Equal(t, "ENTRADA", found.Tipo)
	assert.Equal(t, 5.0, found.Quantidade)
}

func TestMovimentoRepositoryListByProduto(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovimentoRepository(db)

	require.NoError(t, repo.Insert(&models.MovimentoEstoque{ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 5, Origem: "AJUSTE"}))
	require.NoError(t, repo.Insert(&models.MovimentoEstoque{ProdutoID: 1, Tipo: "SAIDA", Quantidade: 2, Origem: "FATURAMENTO"}))
	require.NoError(t, repo.Insert(&models.MovimentoEstoque{ProdutoID: 2, Tipo: "ENTRADA", Quantidade: 1, Origem: "AJUSTE"}))

	movimentos, err := repo.ListByProduto(1)
	require.NoError(t, err)
	assert.Len(t, movimentos, 2)

	movimentos, err = repo.ListByProduto(99)
	require.NoError(t, err)
	assert.Empty(t, movimentos)
}

func TestMovimentoRepositoryFindByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewMovimentoRepository(db)

	_, err := repo.FindByID(999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMapMovimentoErrorUniqueViolation(t *testing.T) {
	err := mapMovimentoError(&pgconn.PgError{Code: "23505"})
	assert.ErrorIs(t, err, ErrDuplicado, "23505 (unique_violation) deve virar ErrDuplicado")
}

func TestMapMovimentoErrorSaldoInsuficiente(t *testing.T) {
	err := mapMovimentoError(&pgconn.PgError{Code: "P0001", Message: "SALDO_INSUFICIENTE"})
	assert.ErrorIs(t, err, ErrSaldoInsuficiente)
}

func TestMapMovimentoErrorProdutoNaoEncontrado(t *testing.T) {
	err := mapMovimentoError(&pgconn.PgError{Code: "P0001", Message: "PRODUTO_NAO_ENCONTRADO"})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMapMovimentoErrorOutroMantido(t *testing.T) {
	origem := errors.New("dial recusado")
	assert.Same(t, origem, mapMovimentoError(origem))
}
