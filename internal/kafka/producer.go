package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexermolov/go-kafka-pusher/internal/config"
	"github.com/segmentio/kafka-go"
)

// Producer handles Kafka message production
type Producer struct {
	writer *kafka.Writer
	cfg    *config.KafkaConfig
	logger *slog.Logger
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig, logger *slog.Logger) (*Producer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config is required")
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		// Topic is now set per-message, not at writer level
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		RequiredAcks: kafka.RequireOne,
		Async:        cfg.Async,
		Compression:  kafka.Snappy,
		Logger:       kafka.LoggerFunc(logger.Debug),
		ErrorLogger:  kafka.LoggerFunc(logger.Error),
	}

	// Use manual partitioning if specific partition is configured
	if cfg.Partition >= 0 {
		writer.Balancer = nil // Manual partition assignment via Message.Partition
	}

	return &Producer{
		writer: writer,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Send sends a message to Kafka
func (p *Producer) Send(ctx context.Context, topic string, message []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Value: message,
		Time:  time.Now(),
	}

	// Set specific partition if configured
	if p.cfg.Partition >= 0 {
		msg.Partition = p.cfg.Partition
	}

	start := time.Now()
	err := p.writer.WriteMessages(ctx, msg)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("failed to send message",
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
		)
		return fmt.Errorf("failed to write message: %w", err)
	}

	p.logger.Info("message sent successfully",
		slog.String("topic", topic),
		slog.Int("size", len(message)),
		slog.Duration("duration", duration),
	)

	return nil
}

// SendBatch sends multiple messages in a batch
func (p *Producer) SendBatch(ctx context.Context, topic string, messages [][]byte) error {
	if len(messages) == 0 {
		return nil
	}

	kafkaMessages := make([]kafka.Message, len(messages))
	for i, msg := range messages {
		kafkaMessages[i] = kafka.Message{
			Topic: topic,
			Value: msg,
			Time:  time.Now(),
		}
		if p.cfg.Partition >= 0 {
			kafkaMessages[i].Partition = p.cfg.Partition
		}
	}

	start := time.Now()
	err := p.writer.WriteMessages(ctx, kafkaMessages...)
	duration := time.Since(start)

	if err != nil {
		p.logger.Error("failed to send batch",
			slog.String("error", err.Error()),
			slog.Int("count", len(messages)),
			slog.Duration("duration", duration),
		)
		return fmt.Errorf("failed to write batch: %w", err)
	}

	p.logger.Info("batch sent successfully",
		slog.String("topic", topic),
		slog.Int("count", len(messages)),
		slog.Duration("duration", duration),
	)

	return nil
}

// Close gracefully closes the producer
func (p *Producer) Close() error {
	if p.writer == nil {
		return nil
	}

	p.logger.Info("closing kafka producer")
	
	if err := p.writer.Close(); err != nil {
		p.logger.Error("failed to close producer",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// Stats returns producer statistics
func (p *Producer) Stats() kafka.WriterStats {
	return p.writer.Stats()
}
