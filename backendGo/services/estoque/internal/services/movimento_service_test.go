package services

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/repositories"
)

type fakeMovimentoRepo struct {
	movimentos []models.MovimentoEstoque
	nextID     uint64
	insertErr  error
}

func (f *fakeMovimentoRepo) Insert(m *models.MovimentoEstoque) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.nextID++
	m.ID = f.nextID
	f.movimentos = append(f.movimentos, *m)
	return nil
}

func (f *fakeMovimentoRepo) FindByID(id uint64) (*models.MovimentoEstoque, error) {
	for i := range f.movimentos {
		if f.movimentos[i].ID == id {
			m := f.movimentos[i]
			return &m, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeMovimentoRepo) ListByProduto(produtoID uint64) ([]models.MovimentoEstoque, error) {
	var result []models.MovimentoEstoque
	for _, m := range f.movimentos {
		if m.ProdutoID == produtoID {
			result = append(result, m)
		}
	}
	return result, nil
}

func newMovimentoServiceForTest(seed ...models.Produto) (MovimentoService, *fakeMovimentoRepo) {
	produtoRepo := newFakeProdutoRepo(seed...)
	movRepo := &fakeMovimentoRepo{}
	return NewMovimentoService(movRepo, produtoRepo), movRepo
}

func TestExecutarEntradaSuccess(t *testing.T) {
	service, movRepo := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	movimento, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 5, Origem: "AJUSTE",
	})

	require.NoError(t, err)
	assert.Equal(t, "ENTRADA", movimento.Tipo)
	assert.Equal(t, 5.0, movimento.Quantidade)
	assert.Len(t, movRepo.movimentos, 1)
}

func TestExecutarSaidaSuccess(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	movimento, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "saida", Quantidade: 3, Origem: "FATURAMENTO", Referencia: "NF-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "SAIDA", movimento.Tipo)
	assert.Equal(t, "NF-1", movimento.Referencia)
}

func TestExecutarSaidaSaldoInsuficiente(t *testing.T) {
	service, movRepo := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 2})

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "SAIDA", Quantidade: 3, Origem: "FATURAMENTO",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
	assert.Empty(t, movRepo.movimentos, "movimento não deve ser gravado quando saldo é insuficiente")
}

func TestExecutarTipoInvalido(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "TRANSFERENCIA", Quantidade: 1, Origem: "AJUSTE",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.BadRequest("").StatusCode, appErr.StatusCode)
}

func TestExecutarQuantidadeZero(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 0, Origem: "AJUSTE",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Unprocessable("").StatusCode, appErr.StatusCode)
}

func TestExecutarOrigemObrigatoria(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 1,
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.BadRequest("").StatusCode, appErr.StatusCode)
}

func TestExecutarProdutoNaoEncontrado(t *testing.T) {
	service, _ := newMovimentoServiceForTest()

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 99, Tipo: "ENTRADA", Quantidade: 1, Origem: "AJUSTE",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestExecutarFalhaBancoMapeadaComoConflito(t *testing.T) {
	produtoRepo := newFakeProdutoRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})
	movRepo := &fakeMovimentoRepo{insertErr: repositories.ErrSaldoInsuficiente}
	service := NewMovimentoService(movRepo, produtoRepo)

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "SAIDA", Quantidade: 1, Origem: "FATURAMENTO",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
}

func TestExecutarMovimentoDuplicadoRetorna409(t *testing.T) {
	produtoRepo := newFakeProdutoRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 10})
	movRepo := &fakeMovimentoRepo{insertErr: repositories.ErrDuplicado}
	service := NewMovimentoService(movRepo, produtoRepo)

	_, err := service.Executar(CreateMovimentoInput{
		ProdutoID: 1, Tipo: "ENTRADA", Quantidade: 1, Origem: "AJUSTE", IdempotencyKey: "chave-1",
	})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.StatusCode)
	assert.Equal(t, apperrors.CodeDuplicado, appErr.Code)
	assert.Contains(t, appErr.Message, "chave-1")
}
