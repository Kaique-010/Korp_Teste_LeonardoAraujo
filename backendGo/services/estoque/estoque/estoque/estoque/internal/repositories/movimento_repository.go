package repositories

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/estoque/internal/models"
)

var (
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
	ErrDuplicado         = errors.New("movimento duplicado (idempotency_key já utilizada)")
)

type MovimentoRepository interface {
	Insert(movimento *models.MovimentoEstoque) error
	FindByID(id uint64) (*models.MovimentoEstoque, error)
	ListByProduto(produtoID uint64) ([]models.MovimentoEstoque, error)
}

type movimentoRepository struct {
	db *gorm.DB
}

func NewMovimentoRepository(db *gorm.DB) MovimentoRepository {
	return &movimentoRepository{db: db}
}

func (r *movimentoRepository) Insert(movimento *models.MovimentoEstoque) error {
	if err := r.db.Create(movimento).Error; err != nil {
		return mapMovimentoError(err)
	}
	return nil
}

func (r *movimentoRepository) FindByID(id uint64) (*models.MovimentoEstoque, error) {
	var movimento models.MovimentoEstoque
	if err := r.db.First(&movimento, id).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return &movimento, nil
}

func (r *movimentoRepository) ListByProduto(produtoID uint64) ([]models.MovimentoEstoque, error) {
	var movimentos []models.MovimentoEstoque
	if err := r.db.Where("produto_id = ?", produtoID).Order("id").Find(&movimentos).Error; err != nil {
		return nil, err
	}
	return movimentos, nil
}

func mapMovimentoError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505": // unique_violation (idempotency_key)
			return ErrDuplicado
		case pgErr.Message == "SALDO_INSUFICIENTE":
			return ErrSaldoInsuficiente
		case pgErr.Message == "PRODUTO_NAO_ENCONTRADO":
			return ErrNotFound
		}
	}
	return err
}
