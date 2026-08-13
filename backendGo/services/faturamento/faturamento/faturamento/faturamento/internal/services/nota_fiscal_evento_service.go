package services

import (
	"encoding/json"

	"gorm.io/datatypes"

	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/repositories"
)

const (
	EventoNotaCriada             = "NOTA_CRIADA"
	EventoItemAdicionado         = "ITEM_ADICIONADO"
	EventoItemRemovido           = "ITEM_REMOVIDO"
	EventoBaixaEstoqueSolicitada = "BAIXA_ESTOQUE_SOLICITADA"
	EventoEstoqueBaixado         = "ESTOQUE_BAIXADO"
	EventoFalhaEstoque           = "FALHA_ESTOQUE"
	EventoNotaFechada            = "NOTA_FECHADA"
)

type NotaFiscalEventoService interface {
	Registrar(notaID uint64, tipo, descricao, referencia string, dados map[string]any) (*models.NotaFiscalEvento, error)
	Listar(notaID uint64) ([]models.NotaFiscalEvento, error)
}

type notaFiscalEventoService struct {
	repo repositories.NotaFiscalEventoRepository
}

func NewNotaFiscalEventoService(repo repositories.NotaFiscalEventoRepository) NotaFiscalEventoService {
	return &notaFiscalEventoService{repo: repo}
}

func (s *notaFiscalEventoService) Registrar(notaID uint64, tipo, descricao, referencia string, dados map[string]any) (*models.NotaFiscalEvento, error) {
	evento := &models.NotaFiscalEvento{
		NotaFiscalID: notaID,
		Tipo:         tipo,
		Descricao:    descricao,
		Referencia:   referencia,
	}

	if dados != nil {
		raw, err := json.Marshal(dados)
		if err != nil {
			return nil, err
		}
		evento.Dados = datatypes.JSON(raw)
	}

	if err := s.repo.Create(evento); err != nil {
		return nil, err
	}
	return evento, nil
}

func (s *notaFiscalEventoService) Listar(notaID uint64) ([]models.NotaFiscalEvento, error) {
	return s.repo.ListByNota(notaID)
}
