package mq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/memodb-io/Acontext/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// tableCarrier adapts amqp.Table to TextMapCarrier for OpenTelemetry propagation
type tableCarrier struct {
	table amqp.Table
}

func (c tableCarrier) Get(key string) string {
	if val, ok := c.table[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func (c tableCarrier) Set(key, value string) {
	c.table[key] = value
}

func (c tableCarrier) Keys() []string {
	keys := make([]string, 0, len(c.table))
	for k := range c.table {
		keys = append(keys, k)
	}
	return keys
}

// DialFunc is a function type for establishing RabbitMQ connections
type DialFunc func() (*amqp.Connection, error)

type Publisher struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	log    *zap.Logger
	cfg    *config.Config
	dialFn DialFunc
	mu     sync.RWMutex
	closed bool
}

type Consumer struct {
	ch  *amqp.Channel
	q   amqp.Queue
	log *zap.Logger
	cfg *config.Config
}

func NewPublisher(conn *amqp.Connection, log *zap.Logger, cfg *config.Config, dialFn DialFunc) (*Publisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Qos(0, 0, false); err != nil {
		return nil, err
	}

	p := &Publisher{
		conn:   conn,
		ch:     ch,
		log:    log,
		cfg:    cfg,
		dialFn: dialFn,
	}

	// Start connection watcher for auto-reconnection
	go p.watchConnection()

	return p, nil
}

// watchConnection monitors the connection and triggers reconnection when closed
func (p *Publisher) watchConnection() {
	for {
		p.mu.RLock()
		if p.closed {
			p.mu.RUnlock()
			return
		}
		conn := p.conn
		p.mu.RUnlock()

		if conn == nil {
			time.Sleep(time.Second)
			continue
		}

		// Wait for connection close notification
		notifyClose := conn.NotifyClose(make(chan *amqp.Error, 1))
		amqpErr := <-notifyClose

		p.mu.RLock()
		if p.closed {
			p.mu.RUnlock()
			return
		}
		p.mu.RUnlock()

		if amqpErr != nil {
			p.log.Warn("RabbitMQ connection closed", zap.Error(amqpErr))
		} else {
			p.log.Warn("RabbitMQ connection closed gracefully")
		}

		// Attempt to reconnect
		p.reconnect()
	}
}

// reconnect attempts to re-establish the RabbitMQ connection with exponential backoff
func (p *Publisher) reconnect() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		p.mu.RLock()
		if p.closed {
			p.mu.RUnlock()
			return
		}
		p.mu.RUnlock()

		p.log.Info("Attempting to reconnect to RabbitMQ", zap.Duration("backoff", backoff))

		conn, err := p.dialFn()
		if err != nil {
			p.log.Error("Failed to reconnect to RabbitMQ", zap.Error(err))
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			p.log.Error("Failed to create channel after reconnect", zap.Error(err))
			conn.Close()
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		if err := ch.Qos(0, 0, false); err != nil {
			p.log.Error("Failed to set QoS after reconnect", zap.Error(err))
			ch.Close()
			conn.Close()
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
			continue
		}

		p.mu.Lock()
		p.conn = conn
		p.ch = ch
		p.mu.Unlock()

		p.log.Info("Successfully reconnected to RabbitMQ")
		return
	}
}

// getChannel safely returns the current channel
func (p *Publisher) getChannel() (*amqp.Channel, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, errors.New("publisher is closed")
	}
	if p.ch == nil {
		return nil, errors.New("channel is not available")
	}
	return p.ch, nil
}

func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	var err error
	if p.ch != nil {
		err = p.ch.Close()
	}
	return err
}

func (p *Publisher) PublishJSON(ctx context.Context, exchangeName string, routingKey string, body any) error {
	b, err := sonic.Marshal(body)
	if err != nil {
		return err
	}

	// Create a span for the publish operation
	tracer := otel.Tracer(p.cfg.App.Name)
	ctx, span := tracer.Start(ctx, "rabbitmq.publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination", exchangeName),
			attribute.String("messaging.destination_kind", "exchange"),
			attribute.String("messaging.rabbitmq.routing_key", routingKey),
		))
	defer span.End()

	// Inject trace context into message headers
	headers := make(amqp.Table)
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, tableCarrier{table: headers})

	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         b,
		Headers:      headers,
	}

	// Get channel safely
	ch, err := p.getChannel()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get channel: %w", err)
	}

	err = ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, publishing)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.SetAttributes(attribute.Int("messaging.message.body.size", len(b)))
	return nil
}

func NewConsumer(conn *amqp.Connection, queueName string, prefetch int, log *zap.Logger, cfg *config.Config) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if prefetch <= 0 {
		prefetch = 10
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return nil, err
	}
	q, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Consumer{ch: ch, q: q, log: log, cfg: cfg}, nil
}

func (c *Consumer) Close() error { return c.ch.Close() }

// Handle is a consumption helper function that will Nack and requeue when the handler returns an error.
func (c *Consumer) Handle(ctx context.Context, handler func([]byte) error) error {
	msgs, err := c.ch.Consume(c.q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	tracer := otel.Tracer(c.cfg.App.Name)
	propagator := otel.GetTextMapPropagator()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-msgs:
			if !ok {
				return errors.New("consumer channel closed")
			}

			// Extract trace context from message headers
			msgCtx := ctx
			if m.Headers != nil {
				msgCtx = propagator.Extract(ctx, tableCarrier{table: m.Headers})
			}

			// Create a span for the consume operation
			// Note: We don't use the returned context since handler doesn't accept context
			_, span := tracer.Start(msgCtx, "rabbitmq.consume",
				trace.WithAttributes(
					attribute.String("messaging.system", "rabbitmq"),
					attribute.String("messaging.destination", c.q.Name),
					attribute.String("messaging.destination_kind", "queue"),
					attribute.String("messaging.operation", "receive"),
					attribute.Int("messaging.message.body.size", len(m.Body)),
				))
			defer span.End()

			// Execute handler with trace context
			// Note: handler receives []byte, not context, so trace context is propagated via span
			if err := handler(m.Body); err != nil {
				span.RecordError(err)
				_ = m.Nack(false, true) // Processing failed, requeue.
				c.log.Sugar().Errorw("consume error", "err", err)
				continue
			}

			_ = m.Ack(false)
		}
	}
}
