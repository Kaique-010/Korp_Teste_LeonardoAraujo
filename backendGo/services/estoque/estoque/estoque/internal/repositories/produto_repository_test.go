package repositories

import (
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/estoque/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&models.Produto{}, &models.MovimentoEstoque{}))
	return db
}

func TestProdutoRepositoryCreateAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	produto := &models.Produto{Codigo: "P001", Descricao: "Caneta", Saldo: 10}
	require.NoError(t, repo.Create(produto))
	assert.NotZero(t, produto.ID)

	found, err := repo.FindByID(produto.ID)
	require.NoError(t, err)
	assert.Equal(t, "P001", found.Codigo)
}

func TestProdutoRepositoryFindByCodigo(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	require.NoError(t, repo.Create(&models.Produto{Codigo: "P001", Descricao: "Caneta"}))

	found, err := repo.FindByCodigo("P001")
	require.NoError(t, err)
	assert.Equal(t, "Caneta", found.Descricao)

	_, err = repo.FindByCodigo("INEXISTENTE")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProdutoRepositoryCodigoUnico(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	require.NoError(t, repo.Create(&models.Produto{Codigo: "P001", Descricao: "Caneta"}))

	err := repo.Create(&models.Produto{Codigo: "P001", Descricao: "Duplicada"})
	assert.Error(t, err, "o índice único deve rejeitar código duplicado")
}

func TestProdutoRepositoryList(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	require.NoError(t, repo.Create(&models.Produto{Codigo: "P001", Descricao: "Caneta"}))
	require.NoError(t, repo.Create(&models.Produto{Codigo: "P002", Descricao: "Caderno"}))

	produtos, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, produtos, 2)
}

func TestProdutoRepositoryUpdate(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	produto := &models.Produto{Codigo: "P001", Descricao: "Caneta", Saldo: 5}
	require.NoError(t, repo.Create(produto))

	produto.Descricao = "Caneta Azul"
	require.NoError(t, repo.Update(produto))

	found, err := repo.FindByID(produto.ID)
	require.NoError(t, err)
	assert.Equal(t, "Caneta Azul", found.Descricao)
}

func TestProdutoRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	produto := &models.Produto{Codigo: "P001", Descricao: "Caneta"}
	require.NoError(t, repo.Create(produto))

	require.NoError(t, repo.Delete(produto.ID))

	_, err := repo.FindByID(produto.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	err = repo.Delete(produto.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestProdutoRepositoryFindByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewProdutoRepository(db)

	_, err := repo.FindByID(999)
	assert.True(t, errors.Is(err, ErrNotFound))
}
