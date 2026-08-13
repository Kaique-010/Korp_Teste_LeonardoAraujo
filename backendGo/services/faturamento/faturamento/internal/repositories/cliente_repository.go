package repositories

import (
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
)

type ClienteRepository interface {
	Create(cliente *models.Cliente) error
	FindByID(id uint64) (*models.Cliente, error)
	List() ([]models.Cliente, error)
	Update(cliente *models.Cliente) error
	Delete(id uint64) error
	TemNotas(clienteID uint64) (bool, error)
}

type clienteRepository struct {
	db *gorm.DB
}

func NewClienteRepository(db *gorm.DB) ClienteRepository {
	return &clienteRepository{db: db}
}

func (r *clienteRepository) Create(c *models.Cliente) error {
	return r.db.Create(c).Error
}

func (r *clienteRepository) FindByID(id uint64) (*models.Cliente, error) {
	var c models.Cliente
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return &c, nil
}

func (r *clienteRepository) List() ([]models.Cliente, error) {
	var clientes []models.Cliente
	if err := r.db.Order("id").Find(&clientes).Error; err != nil {
		return nil, err
	}
	return clientes, nil
}

func (r *clienteRepository) Update(c *models.Cliente) error {
	return r.db.Save(c).Error
}

func (r *clienteRepository) Delete(id uint64) error {
	result := r.db.Delete(&models.Cliente{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *clienteRepository) TemNotas(clienteID uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.NotaFiscal{}).Where("cliente_id = ?", clienteID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
