package repositories

import (
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

// NotaFiscalItemRepository implementa operações de persistência
// para Itens de Nota Fiscal. Separado do NotaFiscalRepository
// para manter responsabilidade única (SPRINT 4 commit 10).
type NotaFiscalItemRepository interface {
	AddItem(item *models.NotaFiscalItem) error
	FindByID(itemID uint64) (*models.NotaFiscalItem, error)
	ListByNota(notaID uint64) ([]models.NotaFiscalItem, error)
	UpdateValorItem(item *models.NotaFiscalItem) error
	DeleteItem(notaID, itemID uint64) error
}

type notaFiscalItemRepository struct {
	db *gorm.DB
}

func NewNotaFiscalItemRepository(db *gorm.DB) NotaFiscalItemRepository {
	return &notaFiscalItemRepository{db: db}
}

func (r *notaFiscalItemRepository) AddItem(item *models.NotaFiscalItem) error {
	return r.db.Create(item).Error
}

func (r *notaFiscalItemRepository) FindByID(itemID uint64) (*models.NotaFiscalItem, error) {
	var it models.NotaFiscalItem
	if err := r.db.First(&it, itemID).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return &it, nil
}

func (r *notaFiscalItemRepository) ListByNota(notaID uint64) ([]models.NotaFiscalItem, error) {
	var itens []models.NotaFiscalItem
	err := r.db.Where("nota_fiscal_id = ?", notaID).Order("id").Find(&itens).Error
	return itens, err
}

func (r *notaFiscalItemRepository) UpdateValorItem(item *models.NotaFiscalItem) error {
	return r.db.Save(item).Error
}

func (r *notaFiscalItemRepository) DeleteItem(notaID, itemID uint64) error {
	res := r.db.Where("nota_fiscal_id = ? AND id = ?", notaID, itemID).
		Delete(&models.NotaFiscalItem{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
