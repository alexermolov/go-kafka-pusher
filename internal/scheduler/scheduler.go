package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/alexermolov/go-kafka-pusher/internal/config"
)

// Task represents a function that will be executed periodically
type Task func(ctx context.Context) error

// Scheduler manages periodic task execution
type Scheduler struct {
	cfg      *config.SchedulerConfig
	logger   *slog.Logger
	task     Task
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  bool
	mu       sync.RWMutex
	stats    Stats
}

// Stats holds scheduler statistics
type Stats struct {
	ExecutionCount uint64
	SuccessCount   uint64
	ErrorCount     uint64
	LastExecution  time.Time
	LastError      error
	mu             sync.RWMutex
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.SchedulerConfig, logger *slog.Logger, task Task) (*Scheduler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("scheduler config is required")
	}
	if task == nil {
		return nil, fmt.Errorf("task is required")
	}
	if cfg.Interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}

	return &Scheduler{
		cfg:    cfg,
		logger: logger,
		task:   task,
	}, nil
}

// Start begins periodic task execution
func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true

	// Create cancellable context
	ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.logger.Info("starting scheduler",
		slog.Duration("interval", s.cfg.Interval),
		slog.Int("workers", s.cfg.WorkerPoolSize),
	)

	// Start worker pool
	taskChan := make(chan struct{}, s.cfg.WorkerPoolSize)
	
	for i := 0; i < s.cfg.WorkerPoolSize; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i, taskChan)
	}

	// Start ticker
	s.wg.Add(1)
	go s.ticker(ctx, taskChan)

	return nil
}

// ticker sends task signals at configured intervals
func (s *Scheduler) ticker(ctx context.Context, taskChan chan<- struct{}) {
	defer s.wg.Done()
	defer close(taskChan)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()

	// Execute immediately on start
	select {
	case taskChan <- struct{}{}:
	case <-ctx.Done():
		return
	}

	for {
		select {
		case <-ticker.C:
			select {
			case taskChan <- struct{}{}:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			s.logger.Info("ticker stopped")
			return
		}
	}
}

// worker executes tasks from the task channel
func (s *Scheduler) worker(ctx context.Context, id int, taskChan <-chan struct{}) {
	defer s.wg.Done()

	s.logger.Debug("worker started", slog.Int("worker_id", id))

	for {
		select {
		case _, ok := <-taskChan:
			if !ok {
				s.logger.Debug("worker stopped", slog.Int("worker_id", id))
				return
			}

			s.executeTask(ctx, id)

		case <-ctx.Done():
			s.logger.Debug("worker stopped by context", slog.Int("worker_id", id))
			return
		}
	}
}

// executeTask executes the configured task and updates statistics
func (s *Scheduler) executeTask(ctx context.Context, workerID int) {
	start := time.Now()

	s.stats.mu.Lock()
	s.stats.ExecutionCount++
	s.stats.LastExecution = start
	s.stats.mu.Unlock()

	s.logger.Debug("executing task",
		slog.Int("worker_id", workerID),
		slog.Uint64("execution", s.stats.ExecutionCount),
	)

	err := s.task(ctx)
	duration := time.Since(start)

	s.stats.mu.Lock()
	if err != nil {
		s.stats.ErrorCount++
		s.stats.LastError = err
		s.logger.Error("task execution failed",
			slog.Int("worker_id", workerID),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
		)
	} else {
		s.stats.SuccessCount++
		s.logger.Info("task executed successfully",
			slog.Int("worker_id", workerID),
			slog.Duration("duration", duration),
		)
	}
	s.stats.mu.Unlock()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is not running")
	}
	s.mu.Unlock()

	s.logger.Info("stopping scheduler")

	// Cancel context if it exists
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}

	// Wait for all workers to finish
	s.wg.Wait()

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.logger.Info("scheduler stopped",
		slog.Uint64("total_executions", s.stats.ExecutionCount),
		slog.Uint64("successful", s.stats.SuccessCount),
		slog.Uint64("failed", s.stats.ErrorCount),
	)

	return nil
}

// IsRunning returns whether the scheduler is currently running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStats returns a copy of current statistics
func (s *Scheduler) GetStats() Stats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()
	
	return Stats{
		ExecutionCount: s.stats.ExecutionCount,
		SuccessCount:   s.stats.SuccessCount,
		ErrorCount:     s.stats.ErrorCount,
		LastExecution:  s.stats.LastExecution,
		LastError:      s.stats.LastError,
	}
}
