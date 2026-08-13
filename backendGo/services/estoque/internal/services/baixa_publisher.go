package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/korp-teste/backendGo/services/estoque/internal/messaging"
)

type resultadoBaixaPublisher struct {
	getChannel ChannelGetter
}

func NewResultadoBaixaPublisher(getChannel ChannelGetter) ResultadoBaixaPublisher {
	return &resultadoBaixaPublisher{getChannel: getChannel}
}

func (p *resultadoBaixaPublisher) PublicarResultado(res messaging.BaixaResultado) error {
	ch := p.getChannel()
	if ch == nil || ch.IsClosed() {
		return fmt.Errorf("canal AMQP indisponível para publicar resultado de baixa")
	}
	body, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("falha ao serializar BaixaResultado: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ch.PublishWithContext(ctx,
		messaging.ExchangeBaixa,
		messaging.KeyBaixaResultado,
		false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}
