package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexermolov/go-kafka-pusher/internal/config"
	"github.com/alexermolov/go-kafka-pusher/internal/kafka"
	"github.com/alexermolov/go-kafka-pusher/internal/logger"
	"github.com/alexermolov/go-kafka-pusher/internal/scheduler"
	"github.com/alexermolov/go-kafka-pusher/internal/template"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "./config.yaml", "path to configuration file")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("kafka-pusher version %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(&cfg.Logging)
	log.Info("starting kafka-pusher",
		slog.String("version", version),
		slog.String("commit", commit),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run application
	if err := run(ctx, cfg, log, sigChan); err != nil {
		log.Error("application error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("kafka-pusher stopped successfully")
}

func run(ctx context.Context, cfg *config.Config, log *slog.Logger, sigChan <-chan os.Signal) error {
	// Initialize template generators for each payload
	type payloadGenerator struct {
		name      string
		generator *template.Generator
		batchSize int
		topic     string
	}

	generators := make([]payloadGenerator, len(cfg.Payloads))
	for i, payloadCfg := range cfg.Payloads {
		gen, err := template.NewGenerator(payloadCfg.TemplatePath)
		if err != nil {
			return fmt.Errorf("failed to create template generator for %s: %w", payloadCfg.Name, err)
		}
		generators[i] = payloadGenerator{
			name:      payloadCfg.Name,
			generator: gen,
			batchSize: payloadCfg.BatchSize,
			topic:     payloadCfg.Topic,
		}
		log.Info("template generator initialized",
			slog.String("name", payloadCfg.Name),
			slog.String("path", payloadCfg.TemplatePath),
			slog.Int("batch_size", payloadCfg.BatchSize),
			slog.String("topic", payloadCfg.Topic),
		)
	}

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(&cfg.Kafka, log)
	if err != nil {
		return fmt.Errorf("failed to create kafka producer: %w", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error("failed to close producer", slog.String("error", err.Error()))
		}
	}()
	log.Info("kafka producer initialized",
		slog.Any("brokers", cfg.Kafka.Brokers),
	)

	// Define the task function
	taskFunc := func(ctx context.Context) error {
		// Process all payloads in parallel
		var wg sync.WaitGroup
		errChan := make(chan error, len(generators))

		for _, pg := range generators {
			wg.Add(1)
			go func(pg payloadGenerator) {
				defer wg.Done()

				// Generate batch of messages from template
				messages := make([][]byte, pg.batchSize)
				for i := 0; i < pg.batchSize; i++ {
					message, err := pg.generator.Generate()
					if err != nil {
						errChan <- fmt.Errorf("failed to generate message %d for %s: %w", i, pg.name, err)
						return
					}
					messages[i] = message

					// Log the message if verbose mode is enabled
					if cfg.Logging.Verbose {
						log.Debug("generated message",
							slog.String("payload", pg.name),
							slog.Int("index", i),
							slog.String("content", string(message)),
						)
					}
				}

				// Send batch to Kafka
				log.Info("sending batch to Kafka",
					slog.String("payload", pg.name),
					slog.String("topic", pg.topic),
					slog.Int("batch_size", len(messages)),
				)
				if err := producer.SendBatch(ctx, pg.topic, messages); err != nil {
					errChan <- fmt.Errorf("failed to send batch for %s: %w", pg.name, err)
					return
				}
			}(pg)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			return err
		}

		return nil
	}

	// If scheduler is enabled, run periodically
	if cfg.Scheduler != nil && cfg.Scheduler.Enabled {
		sched, err := scheduler.NewScheduler(cfg.Scheduler, log, taskFunc)
		if err != nil {
			return fmt.Errorf("failed to create scheduler: %w", err)
		}

		if err := sched.Start(ctx); err != nil {
			return fmt.Errorf("failed to start scheduler: %w", err)
		}
		defer func() {
			if err := sched.Stop(); err != nil {
				log.Error("failed to stop scheduler", slog.String("error", err.Error()))
			}
		}()

		log.Info("scheduler started, waiting for termination signal...")

		// Wait for termination signal
		<-sigChan
		log.Info("received termination signal, shutting down gracefully...")

		// Print statistics
		stats := sched.GetStats()
		log.Info("scheduler statistics",
			slog.Uint64("total_executions", stats.ExecutionCount),
			slog.Uint64("successful", stats.SuccessCount),
			slog.Uint64("failed", stats.ErrorCount),
		)

		return nil
	}

	// Run once if scheduler is not enabled
	log.Info("running in single-shot mode")
	if err := taskFunc(ctx); err != nil {
		return err
	}

	return nil
}
