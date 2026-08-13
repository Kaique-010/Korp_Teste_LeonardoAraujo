package repositories

import (
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

type NotaFiscalEventoRepository interface {
	Create(evento *models.NotaFiscalEvento) error
	ListByNota(notaID uint64) ([]models.NotaFiscalEvento, error)
}

type notaFiscalEventoRepository struct {
	db *gorm.DB
}

func NewNotaFiscalEventoRepository(db *gorm.DB) NotaFiscalEventoRepository {
	return &notaFiscalEventoRepository{db: db}
}

func (r *notaFiscalEventoRepository) Create(evento *models.NotaFiscalEvento) error {
	return r.db.Create(evento).Error
}

func (r *notaFiscalEventoRepository) ListByNota(notaID uint64) ([]models.NotaFiscalEvento, error) {
	var eventos []models.NotaFiscalEvento
	if err := r.db.Where("nota_fiscal_id = ?", notaID).Order("id").Find(&eventos).Error; err != nil {
		return nil, err
	}
	return eventos, nil
}
