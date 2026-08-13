package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/messaging"
	"github.com/korp-teste/backendGo/services/estoque/internal/models"
)

func newBaixaConsumerForTest(movimentos MovimentoService) *baixaConsumer {
	return &baixaConsumer{movimentos: movimentos, publicador: &fakeResultadoBaixaPublisher{}}
}

type fakeResultadoBaixaPublisher struct {
	mensagens []messaging.BaixaResultado
	err       error
}

func (f *fakeResultadoBaixaPublisher) PublicarResultado(msg messaging.BaixaResultado) error {
	f.mensagens = append(f.mensagens, msg)
	return f.err
}

func TestTratarMensagemSucesso(t *testing.T) {
	service, movRepo := newMovimentoServiceForTest(models.Produto{ID: 10, Codigo: "P001", Saldo: 10})
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"numero":7,"itens":[{"produto_id":10,"quantidade":2},{"produto_id":10,"quantidade":3}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, uint64(7), resultado.NotaID)
	assert.Equal(t, messaging.MsgEstoqueBaixado, resultado.Tipo)

	require.Len(t, movRepo.movimentos, 2)
	assert.Equal(t, "SAIDA", movRepo.movimentos[0].Tipo)
	assert.Equal(t, "FATURAMENTO", movRepo.movimentos[0].Origem)
	assert.Equal(t, "nota-7", movRepo.movimentos[0].Referencia)
	assert.Equal(t, "nota-7-item-10", movRepo.movimentos[0].IdempotencyKey)
	assert.Equal(t, float64(2), movRepo.movimentos[0].Quantidade)
	assert.Equal(t, float64(3), movRepo.movimentos[1].Quantidade)
}

func TestTratarMensagemSaldoInsuficiente(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 10, Codigo: "P001", Saldo: 1})
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"itens":[{"produto_id":10,"quantidade":5}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, messaging.MsgBaixaNegada, resultado.Tipo)
	assert.Contains(t, resultado.Motivo, "Saldo insuficiente")
}

func TestTratarMensagemProdutoNaoEncontrado(t *testing.T) {
	service, _ := newMovimentoServiceForTest()
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"itens":[{"produto_id":99,"quantidade":1}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, messaging.MsgBaixaNegada, resultado.Tipo)
	assert.Contains(t, resultado.Motivo, "Produto não encontrado")
}

func TestTratarMensagemQuantidadeInvalida(t *testing.T) {
	service, _ := newMovimentoServiceForTest(models.Produto{ID: 10, Codigo: "P001", Saldo: 10})
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"itens":[{"produto_id":10,"quantidade":0}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, messaging.MsgBaixaNegada, resultado.Tipo)
}

func TestTratarMensagemJsonInvalidoDescartada(t *testing.T) {
	service, _ := newMovimentoServiceForTest()
	consumidor := newBaixaConsumerForTest(service)

	resultado, err := consumidor.tratarMensagem([]byte("isso nao eh json"))

	require.NoError(t, err)
	assert.Nil(t, resultado)
}

func TestTratarMensagemErroTransitorio(t *testing.T) {
	consumidor := newBaixaConsumerForTest(&fakeMovimentoTransiente{})
	body := `{"nota_id":7,"itens":[{"produto_id":10,"quantidade":1}]}`

	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.Error(t, err)
	assert.Nil(t, resultado)
}

func TestEhErroDeNegocio(t *testing.T) {
	assert.True(t, ehErroDeNegocio(apperrors.Conflict("").StatusCode))
	assert.True(t, ehErroDeNegocio(apperrors.NotFound("").StatusCode))
	assert.False(t, ehErroDeNegocio(apperrors.ServiceUnavailable("").StatusCode))
	assert.False(t, ehErroDeNegocio(apperrors.Internal("").StatusCode))
}

type fakeMovimentoTransiente struct{}

func (f *fakeMovimentoTransiente) Executar(input CreateMovimentoInput) (*models.MovimentoEstoque, error) {
	return nil, errors.New("banco indisponível")
}

// fakeMovimentoIdempotente simula o índice único: a 2ª execução da mesma chave
// devolve o mesmo erro 409 que o banco geraria (23505 → ErrDuplicado → Duplicado).
type fakeMovimentoIdempotente struct {
	usadas  map[string]bool
	criados int
}

func (f *fakeMovimentoIdempotente) Executar(input CreateMovimentoInput) (*models.MovimentoEstoque, error) {
	if f.usadas == nil {
		f.usadas = map[string]bool{}
	}
	if f.usadas[input.IdempotencyKey] {
		return nil, apperrors.Duplicado("Movimento já registrado para a chave " + input.IdempotencyKey)
	}
	f.usadas[input.IdempotencyKey] = true
	f.criados++
	return &models.MovimentoEstoque{ID: uint64(f.criados), ProdutoID: input.ProdutoID}, nil
}

func TestTratarMensagemRedeliveryTotalmenteIdempotente(t *testing.T) {
	service := &fakeMovimentoIdempotente{}
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"itens":[{"produto_id":10,"quantidade":2}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))
	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, messaging.MsgEstoqueBaixado, resultado.Tipo)
	assert.Equal(t, 1, service.criados)

	resultado2, err := consumidor.tratarMensagem([]byte(body))
	require.NoError(t, err)
	assert.Nil(t, resultado2, "redelivery com tudo já processado não deve republicar resultado")
	assert.Equal(t, 1, service.criados, "nenhum movimento novo deve ser gravado")
	assert.Empty(t, consumidor.publicador.(*fakeResultadoBaixaPublisher).mensagens,
		"não deve publicar novo resultado para mensagem totalmente duplicada")
}

func TestTratarMensagemRedeliveryParcialIdempotente(t *testing.T) {
	service := &fakeMovimentoIdempotente{usadas: map[string]bool{"nota-7-item-10": true}}
	consumidor := newBaixaConsumerForTest(service)

	body := `{"nota_id":7,"itens":[{"produto_id":10,"quantidade":2},{"produto_id":11,"quantidade":1}]}`
	resultado, err := consumidor.tratarMensagem([]byte(body))

	require.NoError(t, err)
	require.NotNil(t, resultado)
	assert.Equal(t, messaging.MsgEstoqueBaixado, resultado.Tipo,
		"item restante deve ser processado mesmo com item anterior duplicado")
	assert.Equal(t, 1, service.criados, "apenas o item não duplicado deve ser gravado")
}
