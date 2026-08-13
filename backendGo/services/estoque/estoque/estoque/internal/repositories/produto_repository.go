package repositories

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/korp-teste/backendGo/services/estoque/internal/models"
)

var ErrNotFound = errors.New("registro não encontrado")

type ProdutoRepository interface {
	Create(produto *models.Produto) error
	FindByID(id uint64) (*models.Produto, error)
	FindByCodigo(codigo string) (*models.Produto, error)
	List() ([]models.Produto, error)
	Update(produto *models.Produto) error
	Delete(id uint64) error

	TemMovimentos(id uint64) (bool, error)
	ProximoCodigo() (string, error)
	AdicionarPreco(produtoID uint64, precoVista, precoPrazo float64, vigenteEm *time.Time) (*models.PrecoProduto, error)
	PrecoAtual(produtoID uint64) (*models.PrecoProduto, error)
	ListarPrecos(produtoID uint64) ([]models.PrecoProduto, error)
}

type produtoRepository struct {
	db *gorm.DB
}

func NewProdutoRepository(db *gorm.DB) ProdutoRepository {
	return &produtoRepository{db: db}
}

func (r *produtoRepository) Create(produto *models.Produto) error {
	return r.db.Create(produto).Error
}

func (r *produtoRepository) FindByID(id uint64) (*models.Produto, error) {
	var produto models.Produto
	if err := r.db.First(&produto, id).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if err := r.carregarPrecoAtual(&produto); err != nil {
		return nil, err
	}
	return &produto, nil
}

func (r *produtoRepository) FindByCodigo(codigo string) (*models.Produto, error) {
	var produto models.Produto
	if err := r.db.Where("codigo = ?", codigo).First(&produto).Error; err != nil {
		return nil, mapNotFound(err)
	}
	if err := r.carregarPrecoAtual(&produto); err != nil {
		return nil, err
	}
	return &produto, nil
}

func (r *produtoRepository) List() ([]models.Produto, error) {
	var produtos []models.Produto
	if err := r.db.Order("id").Find(&produtos).Error; err != nil {
		return nil, err
	}
	for i := range produtos {
		if err := r.carregarPrecoAtual(&produtos[i]); err != nil {
			return nil, err
		}
	}
	return produtos, nil
}

func (r *produtoRepository) Update(produto *models.Produto) error {
	return r.db.Save(produto).Error
}

func (r *produtoRepository) Delete(id uint64) error {
	result := r.db.Delete(&models.Produto{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *produtoRepository) TemMovimentos(id uint64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.MovimentoEstoque{}).Where("produto_id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *produtoRepository) ProximoCodigo() (string, error) {
	var maxCodigo string
	r.db.Raw("SELECT codigo FROM produtos ORDER BY codigo DESC LIMIT 1").Scan(&maxCodigo)
	if maxCodigo == "" {
		return "PROD-000001", nil
	}

	var numero int
	if _, err := fmt.Sscanf(maxCodigo, "PROD-%d", &numero); err == nil {
		return fmt.Sprintf("PROD-%06d", numero+1), nil
	}

	var maxID uint64
	if err := r.db.Raw("SELECT COALESCE(MAX(id),0) FROM produtos").Scan(&maxID).Error; err != nil {
		return "", err
	}
	return fmt.Sprintf("PROD-%06d", maxID+1), nil
}

func (r *produtoRepository) AdicionarPreco(produtoID uint64, precoVista, precoPrazo float64, vigenteEm *time.Time) (*models.PrecoProduto, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	agora := time.Now()
	inicio := agora
	if vigenteEm != nil {
		inicio = *vigenteEm
	}

	if err := tx.Model(&models.PrecoProduto{}).
		Where("produto_id = ? AND fim_em IS NULL", produtoID).
		Where("vigente_em <= ?", inicio).
		Update("fim_em", inicio).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	novo := &models.PrecoProduto{
		ProdutoID:  produtoID,
		PrecoVista: precoVista,
		PrecoPrazo: precoPrazo,
		VigenteEm:  inicio,
	}
	if err := tx.Create(novo).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return novo, nil
}

func (r *produtoRepository) PrecoAtual(produtoID uint64) (*models.PrecoProduto, error) {
	var preco models.PrecoProduto
	err := r.db.Where("produto_id = ? AND fim_em IS NULL", produtoID).
		Order("vigente_em DESC").
		First(&preco).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &preco, nil
}

func (r *produtoRepository) ListarPrecos(produtoID uint64) ([]models.PrecoProduto, error) {
	var precos []models.PrecoProduto
	if err := r.db.Where("produto_id = ?", produtoID).
		Order("vigente_em DESC").
		Find(&precos).Error; err != nil {
		return nil, err
	}
	return precos, nil
}

func (r *produtoRepository) carregarPrecoAtual(p *models.Produto) error {
	preco, err := r.PrecoAtual(p.ID)
	if err != nil {
		return err
	}
	if preco != nil {
		p.PrecoAtual = &models.PrecoProdutoView{
			PrecoVista: preco.PrecoVista,
			PrecoPrazo: preco.PrecoPrazo,
			VigenteEm:  preco.VigenteEm.Format(time.RFC3339),
		}
	}
	return nil
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
