package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/korp-teste/backendGo/services/estoque/internal/apperrors"
	"github.com/korp-teste/backendGo/services/estoque/internal/logging"
	"github.com/korp-teste/backendGo/services/estoque/internal/messaging"
)

const (
	consumerTag   = "estoque-consumidor-baixa"
	maxTentativas = 5
	backoffBase   = 2 * time.Second
	backoffMax    = 30 * time.Second
)

// BaixaConsumer consome a fila korp.baixa.solicitada, executa as baixas no banco e
// publica o resultado em korp.baixa.resultado.
type BaixaConsumer interface {
	Start() error
	Restart() error
	Close() error
}

type baixaConsumer struct {
	getChannel ChannelGetter
	movimentos MovimentoService
	publicador ResultadoBaixaPublisher
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	cancelCons context.CancelFunc
	mu         sync.Mutex
}

func NewBaixaConsumer(getChannel ChannelGetter, movimentos MovimentoService, publicador ResultadoBaixaPublisher) BaixaConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &baixaConsumer{getChannel: getChannel, movimentos: movimentos, publicador: publicador, ctx: ctx, cancel: cancel}
}

func (c *baixaConsumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.startLocked()
}

func (c *baixaConsumer) Restart() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelCons != nil {
		c.cancelCons()
	}
	c.wg.Wait()
	return c.startLocked()
}

func (c *baixaConsumer) startLocked() error {
	ch := c.getChannel()
	if ch == nil || ch.IsClosed() {
		return fmt.Errorf("canal AMQP indisponível para consumir")
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("falha ao configurar prefetch: %w", err)
	}

	msgs, err := ch.Consume(
		messaging.QueueBaixaSolicitada,
		consumerTag,
		false, // autoAck: false → controle manual (ack/nack)
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("falha ao iniciar consumo de %s: %w", messaging.QueueBaixaSolicitada, err)
	}

	ctx, cancelCons := context.WithCancel(c.ctx)
	c.cancelCons = cancelCons

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.processar(msg)
			}
		}
	}()
	return nil
}

func (c *baixaConsumer) Close() error {
	c.cancel()
	if c.cancelCons != nil {
		c.cancelCons()
	}
	c.wg.Wait()
	ch := c.getChannel()
	if ch != nil && !ch.IsClosed() {
		_ = ch.Cancel(consumerTag, false)
	}
	return nil
}

// processar tenta processar a mensagem com retry e backoff.
// Falha persistente → desiste (ESTOQUE_INDISPONIVEL + DLQ).
func (c *baixaConsumer) processar(msg amqp.Delivery) {
	for tentativa := 1; tentativa <= maxTentativas; tentativa++ {
		resultado, err := c.tratarMensagem(msg.Body)

		if err == nil {
			if resultado != nil {
				if pubErr := c.publicador.PublicarResultado(*resultado); pubErr != nil {
					err = pubErr
				} else {
					_ = msg.Ack(false)
					return
				}
			} else {
				_ = msg.Ack(false)
				return
			}
		}

		if tentativa < maxTentativas {
			logging.Warn("baixa: falha transitória, nova tentativa", map[string]any{
				"tentativa": tentativa,
				"max":       maxTentativas,
				"erro":      err.Error(),
			})
			if !dormirBackoff(c.ctx, tentativa) {
				return
			}
			continue
		}

		c.desistir(msg, err)
		return
	}
}

// desistir avisa o Faturamento (ESTOQUE_INDISPONIVEL) e manda a carta para a DLQ.
func (c *baixaConsumer) desistir(msg amqp.Delivery, causa error) {
	if notaID := notaIDDaMensagem(msg.Body); notaID != 0 {
		_ = c.publicador.PublicarResultado(messaging.BaixaResultado{
			NotaID:     notaID,
			Tipo:       messaging.MsgEstoqueIndisponivel,
			Motivo:     causa.Error(),
			OcorridoEm: time.Now(),
		})
	}
	logging.Error("baixa: desistindo da mensagem após tentativas, enviada para DLQ", map[string]any{
		"max":   maxTentativas,
		"erro":  causa.Error(),
		"nota":  notaIDDaMensagem(msg.Body),
	})
	_ = msg.Nack(false, false) // sem requeue → x-dead-letter → DLQ
}

// tratarMensagem é a lógica pura (sem AMQP), testável.
//
//	*resultado != nil  → publicar resultado e dar ack
//	*resultado == nil  → mensagem descartada (JSON inválido) e dar ack
//	err != nil         → falha transitória (reprocessar depois)
func (c *baixaConsumer) tratarMensagem(body []byte) (*messaging.BaixaResultado, error) {
	var solicitada messaging.BaixaSolicitada
	if err := json.Unmarshal(body, &solicitada); err != nil {
		logging.Warn("baixa: mensagem inválida descartada", map[string]any{"erro": err.Error()})
		return nil, nil
	}

	baixado := false
	for _, item := range solicitada.Itens {
		if _, err := c.movimentos.Executar(CreateMovimentoInput{
			ProdutoID:      item.ProdutoID,
			Tipo:           "SAIDA",
			Quantidade:     item.Quantidade,
			Origem:         "FATURAMENTO",
			Referencia:     fmt.Sprintf("nota-%d", solicitada.NotaID),
			IdempotencyKey: fmt.Sprintf("nota-%d-item-%d", solicitada.NotaID, item.ProdutoID),
		}); err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				if appErr.Code == apperrors.CodeDuplicado {
					logging.Info("baixa: item já baixado (idempotente), ignorado", map[string]any{
						"nota":     solicitada.NotaID,
						"produto":  item.ProdutoID,
					})
					continue
				}
				if ehErroDeNegocio(appErr.StatusCode) {
					return &messaging.BaixaResultado{
						NotaID:     solicitada.NotaID,
						Tipo:       messaging.MsgBaixaNegada,
						Motivo:     appErr.Message,
						OcorridoEm: time.Now(),
					}, nil
				}
			}
			return nil, err // transitório
		}
		baixado = true
	}

	if !baixado {
		logging.Info("baixa: mensagem já totalmente processada (idempotente), descartada", map[string]any{
			"nota": solicitada.NotaID,
		})
		return nil, nil
	}

	return &messaging.BaixaResultado{
		NotaID:     solicitada.NotaID,
		Tipo:       messaging.MsgEstoqueBaixado,
		OcorridoEm: time.Now(),
	}, nil
}

func ehErroDeNegocio(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

func notaIDDaMensagem(body []byte) uint64 {
	var solicitada messaging.BaixaSolicitada
	if err := json.Unmarshal(body, &solicitada); err != nil {
		return 0
	}
	return solicitada.NotaID
}

func dormirBackoff(ctx context.Context, tentativa int) bool {
	d := time.Duration(math.Pow(2, float64(tentativa))) * time.Second
	if d > backoffMax {
		d = backoffMax
	}
	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
