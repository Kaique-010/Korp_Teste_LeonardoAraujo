package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/repositories"
)

const (
	StatusAberta  = "ABERTA"
	StatusFechada = "FECHADA"
)

type CriarNotaInput struct {
	ClienteID *uint64 `json:"cliente_id"`
}

type AdicionarItemInput struct {
	ProdutoID     uint64   `json:"produto_id"`
	Descricao     string   `json:"descricao"`
	Quantidade    float64  `json:"quantidade"`
	ValorUnitario float64  `json:"preco_unitario"`
	PrecoVista    *float64 `json:"preco_vista"`
	PrecoPrazo    *float64 `json:"preco_prazo"`
	Desconto      float64  `json:"desconto"`
}

type NotaFiscalService interface {
	Criar(inputs ...CriarNotaInput) (*models.NotaFiscal, error)
	Obter(id uint64) (*models.NotaFiscal, error)
	Listar() ([]models.NotaFiscal, error)
	AdicionarItem(notaID uint64, input AdicionarItemInput) (*models.NotaFiscal, error)
	RemoverItem(notaID uint64, itemID uint64) error
	Imprimir(notaID uint64) (*models.NotaFiscal, error)
	ProcessarResultadoBaixa(resultado messaging.BaixaResultado) error
	ConfirmarBaixa(notaID uint64, itens []uint64) error
	RejeitarBaixa(notaID uint64, motivo string) error
}

type notaFiscalService struct {
	repo        repositories.NotaFiscalRepository
	clienteRepo repositories.ClienteRepository
	estoque     EstoqueCliente
	eventos     NotaFiscalEventoService
	baixaPub    BaixaPublisher
}

func NewNotaFiscalService(
	repo repositories.NotaFiscalRepository,
	estoque EstoqueCliente,
	eventos NotaFiscalEventoService,
	baixaPub BaixaPublisher,
	clienteRepo ...repositories.ClienteRepository,
) NotaFiscalService {
	s := &notaFiscalService{
		repo:     repo,
		estoque:  estoque,
		eventos:  eventos,
		baixaPub: baixaPub,
	}
	if len(clienteRepo) > 0 {
		s.clienteRepo = clienteRepo[0]
	}
	return s
}

func (s *notaFiscalService) Criar(inputs ...CriarNotaInput) (*models.NotaFiscal, error) {
	var input CriarNotaInput
	if len(inputs) > 0 {
		input = inputs[0]
	}

	numero, err := s.repo.NextNumero()
	if err != nil {
		return nil, err
	}

	var clienteID *uint64
	if input.ClienteID != nil && *input.ClienteID > 0 {
		if s.clienteRepo == nil {
			return nil, apperrors.Internal("Repositório de clientes não configurado")
		}
		if _, err := s.clienteRepo.FindByID(*input.ClienteID); err != nil {
			if errors.Is(err, repositories.ErrNotFound) {
				return nil, apperrors.NotFound("Cliente não encontrado")
			}
			return nil, err
		}
		clienteID = input.ClienteID
	}

	nota := &models.NotaFiscal{
		Numero:    numero,
		Status:    StatusAberta,
		ClienteID: clienteID,
	}
	if err := s.repo.Create(nota); err != nil {
		return nil, err
	}

	if _, err := s.eventos.Registrar(nota.ID, EventoNotaCriada,
		fmt.Sprintf("Nota %d criada", nota.Numero), "", nil); err != nil {
		return nil, err
	}
	return s.repo.FindByID(nota.ID)
}

func (s *notaFiscalService) Obter(id uint64) (*models.NotaFiscal, error) {
	nota, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperrors.NotFound("Nota fiscal não encontrada")
		}
		return nil, err
	}
	return nota, nil
}

func (s *notaFiscalService) Listar() ([]models.NotaFiscal, error) {
	return s.repo.List()
}

func (s *notaFiscalService) AdicionarItem(notaID uint64, input AdicionarItemInput) (*models.NotaFiscal, error) {
	nota, err := s.Obter(notaID)
	if err != nil {
		return nil, err
	}
	if nota.Status != StatusAberta {
		return nil, apperrors.Conflict("Não é possível alterar uma nota já fechada")
	}
	if input.Quantidade <= 0 {
		return nil, apperrors.Unprocessable("Quantidade deve ser positiva")
	}
	if input.ValorUnitario < 0 {
		return nil, apperrors.Unprocessable("Valor unitário não pode ser negativo")
	}
	if input.Desconto < 0 {
		return nil, apperrors.Unprocessable("Desconto não pode ser negativo")
	}

	produto, err := s.estoque.BuscarProduto(input.ProdutoID)
	if err != nil {
		return nil, err
	}

	vista := 0.0
	prazo := 0.0
	if produto.PrecoAtual != nil {
		vista = produto.PrecoAtual.PrecoVista
		prazo = produto.PrecoAtual.PrecoPrazo
	}

	if input.PrecoVista != nil {
		if *input.PrecoVista < 0 {
			return nil, apperrors.Unprocessable("preco_vista não pode ser negativo")
		}
		vista = *input.PrecoVista
	}
	if input.PrecoPrazo != nil {
		if *input.PrecoPrazo < 0 {
			return nil, apperrors.Unprocessable("preco_prazo não pode ser negativo")
		}
		prazo = *input.PrecoPrazo
	}
	valorUnit := input.ValorUnitario
	if valorUnit == 0 && vista > 0 {
		valorUnit = vista
	}

	descricao := strings.TrimSpace(input.Descricao)
	if descricao == "" {
		descricao = produto.Descricao
	}

	total := valorUnit*input.Quantidade - input.Desconto
	if total < 0 {
		return nil, apperrors.Unprocessable("Total não pode ser negativo")
	}

	item := &models.NotaFiscalItem{
		NotaFiscalID:     nota.ID,
		ProdutoID:        produto.ID,
		CodigoProduto:    produto.Codigo,
		DescricaoProduto: descricao,
		Quantidade:       input.Quantidade,
		ValorUnitario:    valorUnit,
		PrecoVista:       vista,
		PrecoPrazo:       prazo,
		Desconto:         input.Desconto,
		Total:            total,
	}
	if err := s.repo.AddItem(item); err != nil {
		return nil, err
	}

	if _, err := s.eventos.Registrar(nota.ID, EventoItemAdicionado,
		fmt.Sprintf("Item %s qtd %g adicionado", item.CodigoProduto, item.Quantidade),
		fmt.Sprintf("item-%d", item.ID), nil); err != nil {
		return nil, err
	}
	return s.repo.FindByID(nota.ID)
}

func (s *notaFiscalService) RemoverItem(notaID uint64, itemID uint64) error {
	nota, err := s.Obter(notaID)
	if err != nil {
		return err
	}
	if nota.Status != StatusAberta {
		return apperrors.Conflict("Não é possível alterar uma nota já fechada")
	}
	if err := s.repo.DeleteItem(notaID, itemID); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return apperrors.NotFound("Item não encontrado na nota")
		}
		return err
	}
	if _, err := s.eventos.Registrar(notaID, EventoItemRemovido,
		fmt.Sprintf("Item %d removido", itemID),
		fmt.Sprintf("item-%d", itemID), nil); err != nil {
		return err
	}
	return nil
}

func (s *notaFiscalService) Imprimir(notaID uint64) (*models.NotaFiscal, error) {
	nota, err := s.Obter(notaID)
	if err != nil {
		return nil, err
	}
	if nota.Status != StatusAberta {
		return nil, apperrors.Conflict("Nota já foi fechada/impressa")
	}
	if len(nota.Itens) == 0 {
		return nil, apperrors.Unprocessable("Nota sem itens não pode ser impressa")
	}

	itens := make([]messaging.ItemBaixa, 0, len(nota.Itens))
	for _, it := range nota.Itens {
		itens = append(itens, messaging.ItemBaixa{
			ProdutoID:  it.ProdutoID,
			Quantidade: it.Quantidade,
		})
	}

	solicitacao := messaging.BaixaSolicitada{
		NotaID: nota.ID,
		Numero: nota.Numero,
		Itens:  itens,
	}
	if err := s.baixaPub.PublicarSolicitacao(solicitacao); err != nil {
		if _, evErr := s.eventos.Registrar(nota.ID, EventoFalhaEstoque,
			"Falha ao publicar solicitação de baixa", "", map[string]any{
				"motivo": err.Error(),
			}); evErr != nil {
			return nil, err
		}
		return nil, apperrors.ServiceUnavailable(fmt.Sprintf("Falha ao enviar solicitação de baixa: %v", err))
	}

	if _, err := s.eventos.Registrar(nota.ID, EventoBaixaEstoqueSolicitada,
		"Baixa de estoque solicitada", fmt.Sprintf("nota-%d", nota.ID), nil); err != nil {
		return nil, err
	}
	return s.repo.FindByID(nota.ID)
}

func (s *notaFiscalService) ProcessarResultadoBaixa(resultado messaging.BaixaResultado) error {
	nota, err := s.repo.FindByID(resultado.NotaID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil
		}
		return err
	}

	switch resultado.Tipo {
	case messaging.MsgEstoqueBaixado:
		if nota.Status == StatusFechada {
			return nil
		}
		nota.Status = StatusFechada
		agora := time.Now()
		nota.FechadoEm = &agora
		if err := s.repo.Update(nota); err != nil {
			return err
		}
		if _, err := s.eventos.Registrar(nota.ID, EventoEstoqueBaixado,
			"Estoque baixado com sucesso", fmt.Sprintf("nota-%d", nota.ID), nil); err != nil {
			return err
		}
		if _, err := s.eventos.Registrar(nota.ID, EventoNotaFechada,
			"Nota fechada após confirmação", fmt.Sprintf("nota-%d", nota.ID), nil); err != nil {
			return err
		}

	case messaging.MsgBaixaNegada, messaging.MsgEstoqueIndisponivel:
		motivo := resultado.Motivo
		if motivo == "" {
			motivo = "Falha na baixa de estoque"
		}
		if _, err := s.eventos.Registrar(nota.ID, EventoFalhaEstoque, motivo,
			fmt.Sprintf("nota-%d", nota.ID), map[string]any{
				"motivo": motivo,
				"tipo":   resultado.Tipo,
			}); err != nil {
			return err
		}

	default:
	}
	return nil
}

func (s *notaFiscalService) ConfirmarBaixa(notaID uint64, itens []uint64) error {
	nota, err := s.Obter(notaID)
	if err != nil {
		return err
	}
	ids := strings.Trim(strings.ReplaceAll(fmt.Sprint(itens), " ", ","), "[]")
	if _, err := s.eventos.Registrar(nota.ID, EventoEstoqueBaixado,
		fmt.Sprintf("Baixa confirmada para itens [%s]", ids), fmt.Sprintf("nota-%d", nota.ID), nil); err != nil {
		return err
	}
	_ = nota
	return nil
}

func (s *notaFiscalService) RejeitarBaixa(notaID uint64, motivo string) error {
	if _, err := s.Obter(notaID); err != nil {
		return err
	}
	mensagem := "Baixa rejeitada"
	if motivo != "" {
		mensagem = fmt.Sprintf("Baixa rejeitada: %s", motivo)
	}
	_, err := s.eventos.Registrar(notaID, EventoFalhaEstoque, mensagem,
		fmt.Sprintf("nota-%d", notaID), map[string]any{"motivo": motivo})
	return err
}
