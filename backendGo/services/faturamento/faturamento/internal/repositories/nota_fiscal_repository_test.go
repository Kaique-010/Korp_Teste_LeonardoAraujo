package repositories

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&models.NotaFiscal{}, &models.NotaFiscalItem{}))
	return db
}

func TestNotaRepositoryCreateAndFindByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	nota := &models.NotaFiscal{Numero: 1, Status: "ABERTA", Itens: []models.NotaFiscalItem{}}
	require.NoError(t, repo.Create(nota))
	assert.NotZero(t, nota.ID)

	found, err := repo.FindByID(nota.ID)
	require.NoError(t, err)
	assert.Equal(t, uint64(1), found.Numero)
	assert.Equal(t, "ABERTA", found.Status)
	assert.Empty(t, found.Itens)
}

func TestNotaRepositoryNumeroUnico(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	require.NoError(t, repo.Create(&models.NotaFiscal{Numero: 1, Status: "ABERTA"}))

	err := repo.Create(&models.NotaFiscal{Numero: 1, Status: "ABERTA"})
	assert.Error(t, err, "o índice único deve rejeitar número duplicado")
}

func TestNotaRepositoryFindByIDComItens(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	nota := &models.NotaFiscal{Numero: 1, Status: "ABERTA"}
	require.NoError(t, repo.Create(nota))

	require.NoError(t, repo.AddItem(&models.NotaFiscalItem{
		NotaFiscalID: nota.ID, ProdutoID: 10, CodigoProduto: "P001",
		DescricaoProduto: "Caneta", Quantidade: 2, ValorUnitario: 10, Total: 20,
	}))
	require.NoError(t, repo.AddItem(&models.NotaFiscalItem{
		NotaFiscalID: nota.ID, ProdutoID: 11, CodigoProduto: "P002",
		DescricaoProduto: "Caderno", Quantidade: 1, ValorUnitario: 5, Total: 5,
	}))

	found, err := repo.FindByID(nota.ID)
	require.NoError(t, err)
	assert.Len(t, found.Itens, 2)
}

func TestNotaRepositoryList(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	require.NoError(t, repo.Create(&models.NotaFiscal{Numero: 2, Status: "ABERTA"}))
	require.NoError(t, repo.Create(&models.NotaFiscal{Numero: 1, Status: "ABERTA"}))

	notas, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, notas, 2)
	assert.Equal(t, uint64(1), notas[0].Numero, "lista ordenada por numero")
}

func TestNotaRepositoryDeleteItem(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	nota := &models.NotaFiscal{Numero: 1, Status: "ABERTA"}
	require.NoError(t, repo.Create(nota))

	item := &models.NotaFiscalItem{NotaFiscalID: nota.ID, ProdutoID: 10, Total: 5}
	require.NoError(t, repo.AddItem(item))

	require.NoError(t, repo.DeleteItem(nota.ID, item.ID))

	err := repo.DeleteItem(nota.ID, item.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestNotaRepositoryFindByIDNotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewNotaFiscalRepository(db)

	_, err := repo.FindByID(999)
	assert.ErrorIs(t, err, ErrNotFound)
}
