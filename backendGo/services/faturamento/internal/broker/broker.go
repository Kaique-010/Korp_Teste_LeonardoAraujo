package broker

import (
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	mu          sync.Mutex
	url         string
	conn        *amqp.Connection
	Channel     *amqp.Channel
	onReconnect func(ch *amqp.Channel) error
}

func Connect(url string, onReconnect ...func(ch *amqp.Channel) error) (*Broker, error) {
	b := &Broker{url: url}
	if len(onReconnect) > 0 {
		b.onReconnect = onReconnect[0]
	}
	if err := b.open(); err != nil {
		return nil, err
	}
	go b.watch()
	return b, nil
}

func (b *Broker) open() error {
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("falha ao conectar no RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("falha ao abrir canal AMQP: %w", err)
	}
	if err := ch.Confirm(false); err != nil {
		conn.Close()
		return fmt.Errorf("falha ao habilitar confirm mode: %w", err)
	}
	if b.onReconnect != nil {
		if err := b.onReconnect(ch); err != nil {
			conn.Close()
			return fmt.Errorf("falha ao executar callback de reconexão: %w", err)
		}
	}
	b.conn = conn
	b.Channel = ch
	return nil
}

func (b *Broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	var firstErr error
	if b.Channel != nil {
		if err := b.Channel.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (b *Broker) Reconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.reconnectLocked()
}

func (b *Broker) reconnectLocked() error {
	_ = b.silentClose()
	for attempt := 1; attempt <= 3; attempt++ {
		if err := b.open(); err != nil {
			time.Sleep(time.Duration(attempt) * 750 * time.Millisecond)
			continue
		}
		return nil
	}
	return errors.New("falha ao reconectar no RabbitMQ após tentativas")
}

func (b *Broker) silentClose() error {
	var firstErr error
	if b.Channel != nil {
		if err := b.Channel.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		b.Channel = nil
	}
	if b.conn != nil {
		if err := b.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		b.conn = nil
	}
	return firstErr
}

func (b *Broker) IsHealthy() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.conn == nil || b.conn.IsClosed() || b.Channel == nil || b.Channel.IsClosed() {
		if err := b.reconnectLocked(); err != nil {
			return fmt.Errorf("conexão AMQP perdida; tentativa de reconexão falhou: %w", err)
		}
	}
	return nil
}

func (b *Broker) ChannelSafe() *amqp.Channel {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Channel
}

func (b *Broker) watch() {
	for {
		b.mu.Lock()
		conn := b.conn
		b.mu.Unlock()
		if conn == nil {
			time.Sleep(2 * time.Second)
			continue
		}
		reason, ok := <-conn.NotifyClose(make(chan *amqp.Error, 1))
		_ = reason
		if !ok {
			return
		}
		b.mu.Lock()
		_ = b.reconnectLocked()
		b.mu.Unlock()
	}
}
