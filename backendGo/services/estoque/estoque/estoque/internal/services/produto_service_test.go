package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/repositories"
)

type fakeProdutoRepo struct {
	produtos []models.Produto
	nextID   uint64
}

func newFakeProdutoRepo(seed ...models.Produto) *fakeProdutoRepo {
	nextID := uint64(0)
	for i := range seed {
		if seed[i].ID > nextID {
			nextID = seed[i].ID
		}
	}
	return &fakeProdutoRepo{produtos: seed, nextID: nextID}
}

func (f *fakeProdutoRepo) Create(p *models.Produto) error {
	f.nextID++
	p.ID = f.nextID
	f.produtos = append(f.produtos, *p)
	return nil
}

func (f *fakeProdutoRepo) FindByID(id uint64) (*models.Produto, error) {
	for i := range f.produtos {
		if f.produtos[i].ID == id {
			p := f.produtos[i]
			return &p, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeProdutoRepo) FindByCodigo(codigo string) (*models.Produto, error) {
	for i := range f.produtos {
		if f.produtos[i].Codigo == codigo {
			p := f.produtos[i]
			return &p, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeProdutoRepo) List() ([]models.Produto, error) {
	return f.produtos, nil
}

func (f *fakeProdutoRepo) Update(p *models.Produto) error {
	for i := range f.produtos {
		if f.produtos[i].ID == p.ID {
			f.produtos[i] = *p
			return nil
		}
	}
	return repositories.ErrNotFound
}

func (f *fakeProdutoRepo) Delete(id uint64) error {
	for i := range f.produtos {
		if f.produtos[i].ID == id {
			f.produtos = append(f.produtos[:i], f.produtos[i+1:]...)
			return nil
		}
	}
	return repositories.ErrNotFound
}

func newProdutoServiceWithFakeRepo(seed ...models.Produto) (ProdutoService, *fakeProdutoRepo) {
	repo := newFakeProdutoRepo(seed...)
	return NewProdutoService(repo), repo
}

func TestCreateProdutoSuccess(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	produto, err := service.Create(CreateProdutoInput{Codigo: "P001", Descricao: "Caneta", Saldo: 10})

	require.NoError(t, err)
	assert.Equal(t, uint64(1), produto.ID)
	assert.Equal(t, "P001", produto.Codigo)
	assert.Equal(t, "Caneta", produto.Descricao)
	assert.Equal(t, 10.0, produto.Saldo)
}

func TestCreateProdutoSaldoDefaultZero(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	produto, err := service.Create(CreateProdutoInput{Codigo: "P001", Descricao: "Caneta"})

	require.NoError(t, err)
	assert.Equal(t, 0.0, produto.Saldo)
}

func TestCreateProdutoCodigoObrigatorio(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	_, err := service.Create(CreateProdutoInput{Descricao: "Caneta"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.BadRequest("").StatusCode, appErr.StatusCode)
}

func TestCreateProdutoDescricaoObrigatoria(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	_, err := service.Create(CreateProdutoInput{Codigo: "P001"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.BadRequest("").StatusCode, appErr.StatusCode)
}

func TestCreateProdutoSaldoNegativo(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	_, err := service.Create(CreateProdutoInput{Codigo: "P001", Descricao: "Caneta", Saldo: -1})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Unprocessable("").StatusCode, appErr.StatusCode)
}

func TestCreateProdutoCodigoDuplicado(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta"})

	_, err := service.Create(CreateProdutoInput{Codigo: "P001", Descricao: "Outra"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
}

func TestGetProdutoNotFound(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	_, err := service.Get(99)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestGetProdutoSuccess(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(models.Produto{ID: 7, Codigo: "P001", Descricao: "Caneta", Saldo: 3})

	produto, err := service.Get(7)

	require.NoError(t, err)
	assert.Equal(t, "P001", produto.Codigo)
	assert.Equal(t, 3.0, produto.Saldo)
}

func TestListProdutos(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(
		models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta"},
		models.Produto{ID: 2, Codigo: "P002", Descricao: "Caderno"},
	)

	produtos, err := service.List()

	require.NoError(t, err)
	assert.Len(t, produtos, 2)
}

func TestUpdateProdutoSuccess(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta", Saldo: 5})

	produto, err := service.Update(1, UpdateProdutoInput{Codigo: "P001", Descricao: "Caneta Azul"})

	require.NoError(t, err)
	assert.Equal(t, "Caneta Azul", produto.Descricao)
	assert.Equal(t, 5.0, produto.Saldo, "saldo não pode mudar via update de cadastro")
}

func TestUpdateProdutoCodigoDuplicado(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(
		models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta"},
		models.Produto{ID: 2, Codigo: "P002", Descricao: "Caderno"},
	)

	_, err := service.Update(1, UpdateProdutoInput{Codigo: "P002", Descricao: "Caneta"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
}

func TestUpdateProdutoMantemProprioCodigo(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta"})

	produto, err := service.Update(1, UpdateProdutoInput{Codigo: "P001", Descricao: "Caneta Azul"})

	require.NoError(t, err)
	assert.Equal(t, "P001", produto.Codigo)
}

func TestUpdateProdutoNotFound(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	_, err := service.Update(99, UpdateProdutoInput{Codigo: "P001", Descricao: "X"})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestDeleteProdutoSuccess(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo(models.Produto{ID: 1, Codigo: "P001", Descricao: "Caneta"})

	require.NoError(t, service.Delete(1))
}

func TestDeleteProdutoNotFound(t *testing.T) {
	service, _ := newProdutoServiceWithFakeRepo()

	err := service.Delete(99)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}
