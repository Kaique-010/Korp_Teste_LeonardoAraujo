package messaging

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Topologia da DLX/DLQ (dead-letter). Cartas que falham demais vão parar aqui.
// Mantido neste arquivo (não no contract.go) porque não faz parte do JSON do contrato.
const (
	ExchangeDlx = "korp.baixa.dlx" // tipo: direct, durável

	QueueBaixaSolicitadaDlq = "korp.baixa.solicitada.dlq"
	KeyBaixaSolicitadaDlq   = "baixa.solicitada.dlq"

	QueueBaixaResultadoDlq = "korp.baixa.resultado.dlq"
	KeyBaixaResultadoDlq   = "baixa.resultado.dlq"
)

// Declare cria a exchange, as filas, os bindings e a DLX/DLQ do fluxo de baixa.
// É idempotente: pode ser chamado por qualquer um dos serviços, a qualquer hora,
// sem risco de erro se a topologia já existir.
func Declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeBaixa, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("falha ao declarar exchange %s: %w", ExchangeBaixa, err)
	}
	if err := ch.ExchangeDeclare(ExchangeDlx, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("falha ao declarar exchange %s: %w", ExchangeDlx, err)
	}

	for _, q := range []struct{ name, key, dlq, dlqKey string }{
		{QueueBaixaSolicitada, KeyBaixaSolicitada, QueueBaixaSolicitadaDlq, KeyBaixaSolicitadaDlq},
		{QueueBaixaResultado, KeyBaixaResultado, QueueBaixaResultadoDlq, KeyBaixaResultadoDlq},
	} {
		args := amqp.Table{
			"x-dead-letter-exchange":    ExchangeDlx,
			"x-dead-letter-routing-key": q.dlqKey,
		}
		if _, err := ch.QueueDeclare(q.name, true, false, false, false, args); err != nil {
			return fmt.Errorf("falha ao declarar fila %s: %w", q.name, err)
		}
		if err := ch.QueueBind(q.name, q.key, ExchangeBaixa, false, nil); err != nil {
			return fmt.Errorf("falha ao ligar fila %s na exchange: %w", q.name, err)
		}

		if _, err := ch.QueueDeclare(q.dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("falha ao declarar fila %s: %w", q.dlq, err)
		}
		if err := ch.QueueBind(q.dlq, q.dlqKey, ExchangeDlx, false, nil); err != nil {
			return fmt.Errorf("falha ao ligar fila %s na DLX: %w", q.dlq, err)
		}
	}
	return nil
}
