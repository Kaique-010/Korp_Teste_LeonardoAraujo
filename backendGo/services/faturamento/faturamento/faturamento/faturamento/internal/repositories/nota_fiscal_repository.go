package repositories

import (
	"errors"

	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

var ErrNotFound = errors.New("registro não encontrado")

type NotaFiscalRepository interface {
	Create(nota *models.NotaFiscal) error
	NextNumero() (uint64, error)
	FindByID(id uint64) (*models.NotaFiscal, error)
	List() ([]models.NotaFiscal, error)
	Update(nota *models.NotaFiscal) error
	AddItem(item *models.NotaFiscalItem) error
	DeleteItem(notaID, itemID uint64) error
	TemItens(notaID uint64) (bool, error)
}

type notaFiscalRepository struct {
	db *gorm.DB
}

func NewNotaFiscalRepository(db *gorm.DB) NotaFiscalRepository {
	return &notaFiscalRepository{db: db}
}

func (r *notaFiscalRepository) Create(nota *models.NotaFiscal) error {
	return r.db.Create(nota).Error
}

func (r *notaFiscalRepository) NextNumero() (uint64, error) {
	var numero uint64
	if err := r.db.Raw("SELECT nextval('nota_fiscal_numero_seq')").Scan(&numero).Error; err != nil {
		return 0, err
	}
	return numero, nil
}

func (r *notaFiscalRepository) FindByID(id uint64) (*models.NotaFiscal, error) {
	var nota models.NotaFiscal
	if err := r.db.Preload("Cliente").Preload("Itens", orderItens).First(&nota, id).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return &nota, nil
}

func (r *notaFiscalRepository) List() ([]models.NotaFiscal, error) {
	var notas []models.NotaFiscal
	if err := r.db.Preload("Cliente").Preload("Itens", orderItens).Order("numero").Find(&notas).Error; err != nil {
		return nil, err
	}
	return notas, nil
}

func (r *notaFiscalRepository) Update(nota *models.NotaFiscal) error {
	return r.db.Save(nota).Error
}

func (r *notaFiscalRepository) AddItem(item *models.NotaFiscalItem) error {
	return r.db.Create(item).Error
}

func (r *notaFiscalRepository) DeleteItem(notaID, itemID uint64) error {
	result := r.db.Where("nota_fiscal_id = ? AND id = ?", notaID, itemID).Delete(&models.NotaFiscalItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *notaFiscalRepository) TemItens(notaID uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.NotaFiscalItem{}).Where("nota_fiscal_id = ?", notaID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func orderItens(db *gorm.DB) *gorm.DB {
	return db.Order("id")
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
