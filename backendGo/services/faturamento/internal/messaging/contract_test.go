package messaging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// O JSON é o contrato: estes testes garantem que os nomes de campo não mudam
// sem querer — os dois serviços precisam trocar exatamente esses JSONs.

func TestBaixaSolicitadaJSON(t *testing.T) {
	msg := BaixaSolicitada{
		NotaID: 7,
		Numero: 1,
		Itens:  []ItemBaixa{{ProdutoID: 10, Quantidade: 2}},
	}

	raw, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"nota_id":7,"numero":1,"itens":[{"produto_id":10,"quantidade":2}]}`,
		string(raw),
	)
}

func TestBaixaResultadoJSON(t *testing.T) {
	ocorrido := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	msg := BaixaResultado{
		NotaID:     7,
		Tipo:       MsgBaixaNegada,
		Motivo:     "saldo insuficiente",
		OcorridoEm: ocorrido,
	}

	raw, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.JSONEq(t,
		`{"nota_id":7,"tipo":"ESTOQUE_BAIXA_NEGADA","motivo":"saldo insuficiente","ocorrido_em":"2026-08-13T12:00:00Z"}`,
		string(raw),
	)
}

func TestBaixaResultadoEstoqueBaixadoSemMotivo(t *testing.T) {
	msg := BaixaResultado{NotaID: 7, Tipo: MsgEstoqueBaixado}

	raw, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "motivo", "motivo não deve aparecer quando vazio")
	assert.Contains(t, string(raw), `"tipo":"ESTOQUE_BAIXADO"`)
}
