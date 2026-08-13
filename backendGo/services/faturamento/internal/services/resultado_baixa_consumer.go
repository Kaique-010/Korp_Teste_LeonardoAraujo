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

	"github.com/korp-teste/backendGo/services/faturamento/internal/apperrors"
	"github.com/korp-teste/backendGo/services/faturamento/internal/logging"
	"github.com/korp-teste/backendGo/services/faturamento/internal/messaging"
)

// Constantes do consumidor de resultado de baixa (Sprint 7 etapa 6).
// - prefetch = 1, autoAck = false (controle total com ack/nack)
// - 5 tentativas com backoff exponencial 2s → 4s → 8s → 15s → 30s
// - 5ª falha: nack(requeue=false) → mensagem segue p/ DLX e a fila DLQ.
const (
	consumerTagResultado = "faturamento-consumidor-resultado-baixa"
	maxTentativasRes     = 5
	backoffBaseRes       = 2 * time.Second
	backoffMaxRes        = 30 * time.Second
)

// ResultadoBaixaConsumer consome a fila korp.baixa.resultado (Estoque → Faturamento)
// e fecha a nota fiscal (ou marca falha) de acordo com o resultado. Garante
// IDEMPOTÊNCIA: se a nota já está FECHADA, apenas confirma a mensagem
// (sem reprocessar, sem publicar nada novo). Sprint 7 etapa 6.
type ResultadoBaixaConsumer interface {
	Start() error
	Restart() error
	Close() error
}

type resultadoBaixaConsumer struct {
	getChannel ChannelGetter
	notas      NotaFiscalService
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	cancelCons context.CancelFunc
	mu         sync.Mutex
}

func NewResultadoBaixaConsumer(getChannel ChannelGetter, notas NotaFiscalService) ResultadoBaixaConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &resultadoBaixaConsumer{getChannel: getChannel, notas: notas, ctx: ctx, cancel: cancel}
}

func (c *resultadoBaixaConsumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.startLocked()
}

func (c *resultadoBaixaConsumer) Restart() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelCons != nil {
		c.cancelCons()
	}
	c.wg.Wait()
	return c.startLocked()
}

// Close termina o consumer de forma graciosa: cancela contexto + espera
// as goroutines terminarem (wg).
func (c *resultadoBaixaConsumer) Close() error {
	c.cancel()
	c.mu.Lock()
	if c.cancelCons != nil {
		c.cancelCons()
	}
	c.mu.Unlock()
	c.wg.Wait()
	return nil
}

// startLocked conecta ao Rabbit, declara QoS e inicia Consume() em goroutine.
// Precisa ser chamado com mu.Lock() adquirido.
func (c *resultadoBaixaConsumer) startLocked() error {
	ch := c.getChannel()
	if ch == nil || ch.IsClosed() {
		return fmt.Errorf("canal AMQP indisponível para consumir resultados de baixa")
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("falha ao definir prefetch no consumidor de resultado: %w", err)
	}
	consCtx, consCancel := context.WithCancel(c.ctx)
	c.cancelCons = consCancel

	msgs, err := ch.Consume(
		messaging.QueueBaixaResultado,
		consumerTagResultado,
		false, // autoAck: NÃO — reconhecemos manualmente após processar
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args extras
	)
	if err != nil {
		consCancel()
		return fmt.Errorf("erro ao iniciar consume da fila %s: %w", messaging.QueueBaixaResultado, err)
	}
	c.wg.Add(1)
	go c.loop(consCtx, msgs)
	return nil
}

// loop recebe os deliveries e processa um a um com retry exponencial.
func (c *resultadoBaixaConsumer) loop(ctx context.Context, msgs <-chan amqp.Delivery) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				logging.Warn("canal de resultados fechado; aguardando reconexão broker")
				return
			}
			c.processarComRetry(ctx, msg)
		}
	}
}

// processarComRetry tenta aplicar ProcessarResultadoBaixa até 5 vezes com
// backoff exponencial. Falhas de negócio (produto inválido etc.) falham
// imediatamente (requeue=false), erros transitórios (DB fora) passam por
// retry. Atingiu 5 → nack(requeue=false) → mensagem vai p/ DLQ.
func (c *resultadoBaixaConsumer) processarComRetry(ctx context.Context, msg amqp.Delivery) {
	var tentativa int
	for {
		tentativa++

		// Backoff APENAS se NÃO for a primeira tentativa
		if tentativa > 1 {
			// exp: 2s, 4s, 8s, 16s → capped 30s
			sleep := time.Duration(math.Pow(2, float64(tentativa-1))) * backoffBaseRes / 2
			if sleep > backoffMaxRes {
				sleep = backoffMaxRes
			}
			select {
			case <-time.After(sleep):
			case <-ctx.Done():
				_ = msg.Nack(false, false)
				return
			}
		}

		err := c.processar(msg)
		if err == nil {
			_ = msg.Ack(false)
			return
		}

		// Erro de negócio / idempotência / nota já fechada?
		// -> NÃO retenta, vai direto ack (não queremos poison).
		if errors.Is(err, apperrors.ErrConflitoEstado) ||
			errors.Is(err, apperrors.ErrNotaJaFechada) ||
			errors.Is(err, errIDempotente) {
			_ = msg.Ack(false)
			return
		}

		logging.Warn("resultado baixa tentativa falhou", map[string]any{
			"tentativa": tentativa,
			"max":       maxTentativasRes,
			"nota_id":   extrairNotaID(msg.Body),
			"err":       err,
		})

		if tentativa >= maxTentativasRes {
			logging.Error("resultado baixa estourou retries → enviando para DLQ", map[string]any{
				"nota_id": extrairNotaID(msg.Body),
				"err":     err,
			})
			_ = msg.Nack(false, false)
			return
		}
	}
}

// errIDempotente é um sentinela interno para casos onde a mensagem já
// foi aplicada e não queremos nem retry nem DLQ.
var errIDempotente = errors.New("idempotente: resultado já aplicado")

// processar parseia a mensagem e delega para NotaFiscalService.
// Importante: essa função é IDEMPOTENTE — se a nota já estiver FECHADA,
// retorna errIDempotente (Ack sem side-effects).
func (c *resultadoBaixaConsumer) processar(msg amqp.Delivery) error {
	var res messaging.BaixaResultado
	if err := json.Unmarshal(msg.Body, &res); err != nil {
		return fmt.Errorf("parse BaixaResultado: %w", err)
	}
	if res.NotaID == 0 {
		return fmt.Errorf("BaixaResultado com nota_id=0 (inválido)")
	}

	// Delegamos toda a lógica de idempotência/fechamento ao service,
	// que também registra os eventos de auditoria (ESTOQUE_BAIXADO /
	// FALHA_ESTOQUE / NOTA_FECHADA) em NotaFiscalEvento JSONB.
	err := c.notas.ProcessarResultadoBaixa(res)
	if err != nil {
		// Mapeamento semântico: service retorna apperrors com tipo + msg;
		// nós elevamos os sentinelas para o loop não retentar.
		var ce *apperrors.AppError
		if errors.As(err, &ce) && ce.Code == "NOTA_JA_FECHADA" {
			return errIDempotente
		}
		return err
	}
	logging.Info("resultado baixa processado", map[string]any{
		"nota_id": res.NotaID,
		"tipo":    res.Tipo,
	})
	return nil
}

// extrairNotaID helper para logs (try parse rápido sem alocar muito).
func extrairNotaID(body []byte) uint64 {
	var r struct {
		NotaID uint64 `json:"nota_id"`
	}
	_ = json.Unmarshal(body, &r)
	return r.NotaID
}
