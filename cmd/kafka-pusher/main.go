package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
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
	// Initialize template generator
	gen, err := template.NewGenerator(cfg.Payload.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to create template generator: %w", err)
	}
	log.Info("template generator initialized", slog.String("path", cfg.Payload.TemplatePath))

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
		slog.String("topic", cfg.Kafka.Topic),
	)

	// Define the task function
	taskFunc := func(ctx context.Context) error {
		// Generate message from template
		message, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate message: %w", err)
		}

		// Log the message if verbose mode is enabled
		if cfg.Logging.Verbose {
			log.Debug("generated message", slog.String("payload", string(message)))
		}

		// Send message to Kafka
		if err := producer.Send(ctx, message); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
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
