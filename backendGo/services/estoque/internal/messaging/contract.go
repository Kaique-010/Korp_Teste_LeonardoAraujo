package messaging

import (
	"time"
)

// ============================================================================
// CONTRATO DE MENSAGERIA — fluxo de baixa de estoque (Sprint 7)
// ============================================================================
// Este arquivo é IDÊNTICO nos dois serviços (faturamento e estoque).
// O JSON trocado é o contrato: qualquer mudança aqui precisa ser replicada
// nos dois. Os testes contract_test.go travam o formato.
// ============================================================================

// Tipos de mensagem (campo "tipo" usado nas cartas de resultado).
const (
	MsgBaixaSolicitada     = "ESTOQUE_BAIXA_SOLICITADA"
	MsgEstoqueBaixado      = "ESTOQUE_BAIXADO"
	MsgBaixaNegada         = "ESTOQUE_BAIXA_NEGADA"
	MsgEstoqueIndisponivel = "ESTOQUE_INDISPONIVEL"
)

// Topologia: UMA exchange direct com DUAS filas.
// A carta entra na exchange e o binding (routing key) decide em qual fila cai.
const (
	ExchangeBaixa = "korp.baixa" // tipo: direct, durável

	QueueBaixaSolicitada = "korp.baixa.solicitada" // quem consome: Estoque
	KeyBaixaSolicitada   = "baixa.solicitada"

	QueueBaixaResultado = "korp.baixa.resultado" // quem consome: Faturamento
	KeyBaixaResultado   = "baixa.resultado"
)

// ItemBaixa representa um produto e a quantidade a ser baixada.
type ItemBaixa struct {
	ProdutoID  uint64  `json:"produto_id"`
	Quantidade float64 `json:"quantidade"`
}

// BaixaSolicitada é a carta Faturamento → Estoque.
type BaixaSolicitada struct {
	NotaID uint64      `json:"nota_id"`
	Numero uint64      `json:"numero"`
	Itens  []ItemBaixa `json:"itens"`
}

// BaixaResultado é a carta Estoque → Faturamento.
type BaixaResultado struct {
	NotaID     uint64    `json:"nota_id"`
	Tipo       string    `json:"tipo"` // ESTOQUE_BAIXADO | ESTOQUE_BAIXA_NEGADA
	Motivo     string    `json:"motivo,omitempty"`
	OcorridoEm time.Time `json:"ocorrido_em"`
}
