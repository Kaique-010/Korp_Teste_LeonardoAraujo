package services

import (
	"errors"
	"strings"
	"time"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/repositories"
)

type CreateProdutoInput struct {
	Codigo     string   `json:"codigo"`
	Descricao  string   `json:"descricao"`
	Saldo      float64  `json:"saldo"`
	PrecoVista *float64 `json:"preco_vista"`
	PrecoPrazo *float64 `json:"preco_prazo"`
	VigenteEm  *string  `json:"vigente_em"`
}

type UpdateProdutoInput struct {
	Codigo     string   `json:"codigo"`
	Descricao  string   `json:"descricao"`
	PrecoVista *float64 `json:"preco_vista"`
	PrecoPrazo *float64 `json:"preco_prazo"`
	VigenteEm  *string  `json:"vigente_em"`
}

type AtualizarPrecoInput struct {
	PrecoVista float64 `json:"preco_vista"`
	PrecoPrazo float64 `json:"preco_prazo"`
	VigenteEm  *string `json:"vigente_em"`
}

type ProdutoService interface {
	Create(input CreateProdutoInput) (*models.Produto, error)
	Get(id uint64) (*models.Produto, error)
	List() ([]models.Produto, error)
	Update(id uint64, input UpdateProdutoInput) (*models.Produto, error)
	Delete(id uint64) error

	ListarPrecos(produtoID uint64) ([]models.PrecoProduto, error)
	AtualizarPreco(produtoID uint64, input AtualizarPrecoInput) (*models.PrecoProduto, error)
}

type produtoService struct {
	repo repositories.ProdutoRepository
}

func NewProdutoService(repo repositories.ProdutoRepository) ProdutoService {
	return &produtoService{repo: repo}
}

func (s *produtoService) Create(input CreateProdutoInput) (*models.Produto, error) {
	input.Codigo = strings.TrimSpace(input.Codigo)
	input.Descricao = strings.TrimSpace(input.Descricao)

	if input.Codigo == "" {
		prox, err := s.repo.ProximoCodigo()
		if err != nil {
			return nil, err
		}
		input.Codigo = prox
	}
	if input.Descricao == "" {
		return nil, apperrors.BadRequest("Descrição é obrigatória")
	}
	if input.Saldo < 0 {
		return nil, apperrors.Unprocessable("Saldo não pode ser negativo")
	}
	if (input.PrecoVista != nil && *input.PrecoVista < 0) ||
		(input.PrecoPrazo != nil && *input.PrecoPrazo < 0) {
		return nil, apperrors.Unprocessable("Preços não podem ser negativos")
	}

	if err := s.ensureCodigoDisponivel(input.Codigo); err != nil {
		return nil, err
	}

	produto := &models.Produto{
		Codigo:    input.Codigo,
		Descricao: input.Descricao,
		Saldo:     input.Saldo,
	}
	if err := s.repo.Create(produto); err != nil {
		return nil, err
	}

	if input.PrecoVista != nil || input.PrecoPrazo != nil {
		vig, err := parseVigencia(input.VigenteEm)
		if err != nil {
			return nil, err
		}
		vista := 0.0
		if input.PrecoVista != nil {
			vista = *input.PrecoVista
		}
		prazo := 0.0
		if input.PrecoPrazo != nil {
			prazo = *input.PrecoPrazo
		}
		if _, err := s.repo.AdicionarPreco(produto.ID, vista, prazo, vig); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(produto.ID)
}

func (s *produtoService) Get(id uint64) (*models.Produto, error) {
	produto, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Produto não encontrado")
		}
		return nil, err
	}
	return produto, nil
}

func (s *produtoService) List() ([]models.Produto, error) {
	return s.repo.List()
}

func (s *produtoService) Update(id uint64, input UpdateProdutoInput) (*models.Produto, error) {
	input.Codigo = strings.TrimSpace(input.Codigo)
	input.Descricao = strings.TrimSpace(input.Descricao)

	if input.Codigo == "" {
		return nil, apperrors.BadRequest("Código é obrigatório")
	}
	if input.Descricao == "" {
		return nil, apperrors.BadRequest("Descrição é obrigatória")
	}
	if (input.PrecoVista != nil && *input.PrecoVista < 0) ||
		(input.PrecoPrazo != nil && *input.PrecoPrazo < 0) {
		return nil, apperrors.Unprocessable("Preços não podem ser negativos")
	}

	produto, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Produto não encontrado")
		}
		return nil, err
	}

	if input.Codigo != produto.Codigo {
		if err := s.ensureCodigoDisponivel(input.Codigo); err != nil {
			return nil, err
		}
	}

	produto.Codigo = input.Codigo
	produto.Descricao = input.Descricao

	if input.PrecoVista != nil || input.PrecoPrazo != nil {
		atual, err := s.repo.PrecoAtual(id)
		if err != nil {
			return nil, err
		}
		vista := 0.0
		prazo := 0.0
		if atual != nil {
			vista = atual.PrecoVista
			prazo = atual.PrecoPrazo
		}
		if input.PrecoVista != nil {
			vista = *input.PrecoVista
		}
		if input.PrecoPrazo != nil {
			prazo = *input.PrecoPrazo
		}
		vig, err := parseVigencia(input.VigenteEm)
		if err != nil {
			return nil, err
		}
		if _, err := s.repo.AdicionarPreco(id, vista, prazo, vig); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(produto); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *produtoService) Delete(id uint64) error {
	if _, err := s.repo.FindByID(id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return apperrors.NotFound("Produto não encontrado")
		}
		return err
	}
	tem, err := s.repo.TemMovimentos(id)
	if err != nil {
		return err
	}
	if tem {
		return apperrors.Conflict("Não é possível excluir produto com movimentos de estoque")
	}
	return s.repo.Delete(id)
}

func (s *produtoService) ListarPrecos(produtoID uint64) ([]models.PrecoProduto, error) {
	if _, err := s.repo.FindByID(produtoID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Produto não encontrado")
		}
		return nil, err
	}
	return s.repo.ListarPrecos(produtoID)
}

func (s *produtoService) AtualizarPreco(produtoID uint64, input AtualizarPrecoInput) (*models.PrecoProduto, error) {
	if _, err := s.repo.FindByID(produtoID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Produto não encontrado")
		}
		return nil, err
	}
	if input.PrecoVista < 0 || input.PrecoPrazo < 0 {
		return nil, apperrors.Unprocessable("Preços não podem ser negativos")
	}
	vig, err := parseVigencia(input.VigenteEm)
	if err != nil {
		return nil, err
	}
	return s.repo.AdicionarPreco(produtoID, input.PrecoVista, input.PrecoPrazo, vig)
}

func (s *produtoService) ensureCodigoDisponivel(codigo string) error {
	_, err := s.repo.FindByCodigo(codigo)
	if err == nil {
		return apperrors.Conflict("Já existe um produto com esse código")
	}
	if !errors.Is(err, repositories.ErrNotFound) {
		return err
	}
	return nil
}

func parseVigencia(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02", *s); err2 == nil {
			return &t2, nil
		}
		return nil, apperrors.BadRequest("Formato de data inválido para vigência (use RFC3339 ou YYYY-MM-DD)")
	}
	return &t, nil
}
