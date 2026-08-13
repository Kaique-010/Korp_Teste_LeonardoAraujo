package services

import (
	"errors"
	"strings"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
	"github.com/korp-teste/backendGo/services/estoque/internal/repositories"
)

type CreateMovimentoInput struct {
	ProdutoID      uint64  `json:"produto_id"`
	Tipo           string  `json:"tipo"`
	Quantidade     float64 `json:"quantidade"`
	Origem         string  `json:"origem"`
	Referencia     string  `json:"referencia"`
	IdempotencyKey string  `json:"idempotency_key"`
}

type MovimentoService interface {
	Executar(input CreateMovimentoInput) (*models.MovimentoEstoque, error)
}

type movimentoService struct {
	movimentos repositories.MovimentoRepository
	produtos   repositories.ProdutoRepository
}

func NewMovimentoService(movimentos repositories.MovimentoRepository, produtos repositories.ProdutoRepository) MovimentoService {
	return &movimentoService{movimentos: movimentos, produtos: produtos}
}

func (s *movimentoService) Executar(input CreateMovimentoInput) (*models.MovimentoEstoque, error) {
	input.Tipo = strings.ToUpper(strings.TrimSpace(input.Tipo))
	input.Origem = strings.TrimSpace(input.Origem)

	if input.Tipo != "ENTRADA" && input.Tipo != "SAIDA" {
		return nil, apperrors.BadRequest("Tipo de movimento deve ser ENTRADA ou SAIDA")
	}
	if input.Quantidade <= 0 {
		return nil, apperrors.Unprocessable("Quantidade deve ser maior que zero")
	}
	if input.Origem == "" {
		return nil, apperrors.BadRequest("Origem é obrigatória")
	}

	produto, err := s.produtos.FindByID(input.ProdutoID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Produto não encontrado")
		}
		return nil, err
	}

	if input.Tipo == "SAIDA" && produto.Saldo < input.Quantidade {
		return nil, apperrors.Conflict("Saldo insuficiente para o produto " + produto.Codigo)
	}

	movimento := &models.MovimentoEstoque{
		ProdutoID:      input.ProdutoID,
		Tipo:           input.Tipo,
		Quantidade:     input.Quantidade,
		Origem:         input.Origem,
		Referencia:     input.Referencia,
		IdempotencyKey: input.IdempotencyKey,
	}

	if err := s.movimentos.Insert(movimento); err != nil {
		switch {
		case errors.Is(err, repositories.ErrSaldoInsuficiente):
			return nil, apperrors.Conflict("Saldo insuficiente para o produto " + produto.Codigo)
		case errors.Is(err, repositories.ErrNotFound):
			return nil, apperrors.NotFound("Produto não encontrado")
		case errors.Is(err, repositories.ErrDuplicado):
			return nil, apperrors.Duplicado("Movimento já registrado para a chave " + input.IdempotencyKey)
		}
		return nil, err
	}
	return movimento, nil
}
