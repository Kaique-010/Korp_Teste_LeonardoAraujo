package services

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/models"
	"github.com/korp-teste/backendGo/services/faturamento/internal/repositories"
)

type fakeNotaRepo struct {
	notas      []models.NotaFiscal
	itens      []models.NotaFiscalItem
	nextNota   uint64
	nextItem   uint64
	nextNum    uint64
	falhaUpdate bool
}

func newFakeNotaRepo() *fakeNotaRepo {
	return &fakeNotaRepo{}
}

func (f *fakeNotaRepo) Create(n *models.NotaFiscal) error {
	f.nextNota++
	n.ID = f.nextNota
	f.notas = append(f.notas, *n)
	return nil
}

func (f *fakeNotaRepo) NextNumero() (uint64, error) {
	f.nextNum++
	return f.nextNum, nil
}

func (f *fakeNotaRepo) FindByID(id uint64) (*models.NotaFiscal, error) {
	for i := range f.notas {
		if f.notas[i].ID == id {
			nota := f.notas[i]
			for _, item := range f.itens {
				if item.NotaFiscalID == id {
					nota.Itens = append(nota.Itens, item)
				}
			}
			if nota.Itens == nil {
				nota.Itens = []models.NotaFiscalItem{}
			}
			return &nota, nil
		}
	}
	return nil, repositories.ErrNotFound
}

func (f *fakeNotaRepo) List() ([]models.NotaFiscal, error) {
	return f.notas, nil
}

func (f *fakeNotaRepo) Update(n *models.NotaFiscal) error {
	if f.falhaUpdate {
		return errors.New("banco indisponível")
	}
	for i := range f.notas {
		if f.notas[i].ID == n.ID {
			f.notas[i] = *n
			return nil
		}
	}
	return repositories.ErrNotFound
}

func (f *fakeNotaRepo) AddItem(item *models.NotaFiscalItem) error {
	f.nextItem++
	item.ID = f.nextItem
	f.itens = append(f.itens, *item)
	return nil
}

func (f *fakeNotaRepo) DeleteItem(notaID, itemID uint64) error {
	for i := range f.itens {
		if f.itens[i].ID == itemID && f.itens[i].NotaFiscalID == notaID {
			f.itens = append(f.itens[:i], f.itens[i+1:]...)
			return nil
		}
	}
	return repositories.ErrNotFound
}

type fakeEstoqueCliente struct {
	produto  *EstoqueProduto
	err      error
	baixa    []BaixaEstoqueInput
	baixaErr error
}

func (f *fakeEstoqueCliente) BuscarProduto(produtoID uint64) (*EstoqueProduto, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.produto, nil
}

func (f *fakeEstoqueCliente) SolicitarBaixa(input BaixaEstoqueInput) error {
	f.baixa = append(f.baixa, input)
	if f.baixaErr != nil {
		return f.baixaErr
	}
	return nil
}

type fakeNotaEventoService struct {
	registrados []*models.NotaFiscalEvento
	err         error
}

func (f *fakeNotaEventoService) Registrar(notaID uint64, tipo, descricao, referencia string, dados map[string]any) (*models.NotaFiscalEvento, error) {
	if f.err != nil {
		return nil, f.err
	}
	e := &models.NotaFiscalEvento{NotaFiscalID: notaID, Tipo: tipo, Descricao: descricao, Referencia: referencia}
	if dados != nil {
		raw, _ := json.Marshal(dados)
		e.Dados = datatypes.JSON(raw)
	}
	f.registrados = append(f.registrados, e)
	return e, nil
}

func (f *fakeNotaEventoService) Listar(notaID uint64) ([]models.NotaFiscalEvento, error) {
	return nil, nil
}

type fakeBaixaPublisher struct {
	mensagens []messaging.BaixaSolicitada
	err       error
}

func (f *fakeBaixaPublisher) PublicarSolicitacao(msg messaging.BaixaSolicitada) error {
	f.mensagens = append(f.mensagens, msg)
	return f.err
}

func seedAberta(repo *fakeNotaRepo) *models.NotaFiscal {
	nota := &models.NotaFiscal{Numero: 1, Status: StatusAberta}
	repo.Create(nota)
	return nota
}

func newNotaService(repo *fakeNotaRepo, estoque *fakeEstoqueCliente) NotaFiscalService {
	return NewNotaFiscalService(repo, estoque, &fakeNotaEventoService{}, &fakeBaixaPublisher{})
}

func newNotaServiceComEventos(repo *fakeNotaRepo, estoque *fakeEstoqueCliente, eventos *fakeNotaEventoService) NotaFiscalService {
	return NewNotaFiscalService(repo, estoque, eventos, &fakeBaixaPublisher{})
}

func newNotaServiceComDeps(repo *fakeNotaRepo, estoque *fakeEstoqueCliente, eventos *fakeNotaEventoService, publicador *fakeBaixaPublisher) NotaFiscalService {
	return NewNotaFiscalService(repo, estoque, eventos, publicador)
}

func TestCriarNota(t *testing.T) {
	repo := newFakeNotaRepo()
	service := newNotaService(repo, &fakeEstoqueCliente{})

	nota, err := service.Criar()

	require.NoError(t, err)
	assert.Equal(t, uint64(1), nota.Numero)
	assert.Equal(t, StatusAberta, nota.Status)
	assert.Empty(t, nota.Itens)
}

func TestObterNotaNotFound(t *testing.T) {
	repo := newFakeNotaRepo()
	service := newNotaService(repo, &fakeEstoqueCliente{})

	_, err := service.Obter(99)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemSuccess(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{
		produto: &EstoqueProduto{ID: 10, Codigo: "P001", Descricao: "Caneta", Saldo: 5},
	})

	updated, err := service.AdicionarItem(nota.ID, AdicionarItemInput{
		ProdutoID: 10, Quantidade: 3, ValorUnitario: 10, Desconto: 5,
	})

	require.NoError(t, err)
	require.Len(t, updated.Itens, 1)
	item := updated.Itens[0]
	assert.Equal(t, "P001", item.CodigoProduto, "snapshot do código")
	assert.Equal(t, "Caneta", item.DescricaoProduto, "snapshot da descrição")
	assert.Equal(t, 25.0, item.Total, "3*10 - 5")
}

func TestAdicionarItemNotaFechada(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	nota.Status = StatusFechada
	repo.Update(nota)
	service := newNotaService(repo, &fakeEstoqueCliente{})

	_, err := service.AdicionarItem(nota.ID, AdicionarItemInput{ProdutoID: 10, Quantidade: 1, ValorUnitario: 5})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Conflict("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemNotaNotFound(t *testing.T) {
	repo := newFakeNotaRepo()
	service := newNotaService(repo, &fakeEstoqueCliente{})

	_, err := service.AdicionarItem(99, AdicionarItemInput{ProdutoID: 10, Quantidade: 1, ValorUnitario: 5})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemProdutoNaoExiste(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{
		err: apperrors.NotFound("Produto não encontrado no Serviço de Estoque"),
	})

	_, err := service.AdicionarItem(nota.ID, AdicionarItemInput{ProdutoID: 99, Quantidade: 1, ValorUnitario: 5})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemEstoqueIndisponivel(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{
		err: apperrors.ServiceUnavailable("Serviço de Estoque indisponível"),
	})

	_, err := service.AdicionarItem(nota.ID, AdicionarItemInput{ProdutoID: 10, Quantidade: 1, ValorUnitario: 5})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.ServiceUnavailable("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemQuantidadeInvalida(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{produto: &EstoqueProduto{ID: 10, Codigo: "P001", Descricao: "Caneta"}})

	_, err := service.AdicionarItem(nota.ID, AdicionarItemInput{ProdutoID: 10, Quantidade: 0, ValorUnitario: 5})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Unprocessable("").StatusCode, appErr.StatusCode)
}

func TestAdicionarItemDescontoExcedeTotal(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{produto: &EstoqueProduto{ID: 10, Codigo: "P001", Descricao: "Caneta"}})

	_, err := service.AdicionarItem(nota.ID, AdicionarItemInput{ProdutoID: 10, Quantidade: 1, ValorUnitario: 5, Desconto: 10})

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.Unprocessable("").StatusCode, appErr.StatusCode)
}

func TestRemoverItemSuccess(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	item := &models.NotaFiscalItem{NotaFiscalID: nota.ID, ProdutoID: 10, Total: 5}
	repo.AddItem(item)

	service := newNotaService(repo, &fakeEstoqueCliente{})

	require.NoError(t, service.RemoverItem(nota.ID, item.ID))
}

func TestRemoverItemNaoEncontrado(t *testing.T) {
	repo := newFakeNotaRepo()
	nota := seedAberta(repo)
	service := newNotaService(repo, &fakeEstoqueCliente{})

	err := service.RemoverItem(nota.ID, 999)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperrors.NotFound("").StatusCode, appErr.StatusCode)
}
