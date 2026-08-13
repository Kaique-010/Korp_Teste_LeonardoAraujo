package services

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
)

// ChannelGetter retorna o canal AMQP atual (pode trocar após reconexão).
type ChannelGetter func() *amqp.Channel

// BaixaPublisher publica a carta de baixa de estoque na fila do RabbitMQ.
type BaixaPublisher interface {
	PublicarSolicitacao(msg messaging.BaixaSolicitada) error
}

type amqpBaixaPublisher struct {
	getChannel ChannelGetter
}

func NewBaixaPublisher(getChannel ChannelGetter) BaixaPublisher {
	return &amqpBaixaPublisher{getChannel: getChannel}
}

func (p *amqpBaixaPublisher) PublicarSolicitacao(msg messaging.BaixaSolicitada) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("falha ao serializar mensagem de baixa: %w", err)
	}
	ch := p.getChannel()
	if ch == nil || ch.IsClosed() {
		return fmt.Errorf("canal AMQP indisponível")
	}
	return ch.Publish(
		messaging.ExchangeBaixa,
		messaging.KeyBaixaSolicitada,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
