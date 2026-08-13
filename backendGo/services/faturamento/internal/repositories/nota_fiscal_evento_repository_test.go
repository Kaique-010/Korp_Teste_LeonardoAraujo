package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

func TestNotaFiscalEventoRepositoryCreateELista(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.NotaFiscalEvento{}))
	repo := NewNotaFiscalEventoRepository(db)

	require.NoError(t, repo.Create(&models.NotaFiscalEvento{
		NotaFiscalID: 1,
		Tipo:         "NOTA_CRIADA",
		Descricao:    "Nota fiscal criada",
		Dados:        datatypes.JSON([]byte(`{"numero":1}`)),
	}))
	require.NoError(t, repo.Create(&models.NotaFiscalEvento{
		NotaFiscalID: 1,
		Tipo:         "ITEM_ADICIONADO",
		Descricao:    "Item adicionado",
	}))

	eventos, err := repo.ListByNota(1)
	require.NoError(t, err)
	require.Len(t, eventos, 2)
	assert.Equal(t, "NOTA_CRIADA", eventos[0].Tipo)
	assert.Equal(t, "ITEM_ADICIONADO", eventos[1].Tipo)
	assert.JSONEq(t, `{"numero":1}`, string(eventos[0].Dados))
}

func TestNotaFiscalEventoRepositoryListaPorNota(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, db.AutoMigrate(&models.NotaFiscalEvento{}))
	repo := NewNotaFiscalEventoRepository(db)

	require.NoError(t, repo.Create(&models.NotaFiscalEvento{NotaFiscalID: 1, Tipo: "NOTA_CRIADA", Descricao: "x"}))
	require.NoError(t, repo.Create(&models.NotaFiscalEvento{NotaFiscalID: 2, Tipo: "NOTA_CRIADA", Descricao: "y"}))

	eventos, err := repo.ListByNota(1)
	require.NoError(t, err)
	require.Len(t, eventos, 1)
	assert.Equal(t, uint64(1), eventos[0].NotaFiscalID)
}
