package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Kafka     KafkaConfig     `yaml:"kafka" validate:"required"`
	Scheduler *SchedulerConfig `yaml:"scheduler,omitempty"`
	Logging   LoggingConfig   `yaml:"logging"`
	Payload   PayloadConfig   `yaml:"payload" validate:"required"`
}

// KafkaConfig holds Kafka connection settings
type KafkaConfig struct {
	Brokers   []string      `yaml:"brokers" validate:"required,min=1"`
	Topic     string        `yaml:"topic" validate:"required"`
	ClientID  string        `yaml:"client_id"`
	Partition int           `yaml:"partition"`
	Timeout   time.Duration `yaml:"timeout"`
	BatchSize int           `yaml:"batch_size"`
	Async     bool          `yaml:"async"`
}

// SchedulerConfig holds scheduler settings
type SchedulerConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Interval       time.Duration `yaml:"interval" validate:"required_if=Enabled true"`
	WorkerPoolSize int           `yaml:"worker_pool_size"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level   string `yaml:"level"`
	Format  string `yaml:"format"` // json or text
	Verbose bool   `yaml:"verbose"`
}

// PayloadConfig holds payload template settings
type PayloadConfig struct {
	TemplatePath string `yaml:"template_path" validate:"required"`
	BatchSize    int    `yaml:"batch_size"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	cfg.setDefaults()

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default values for optional fields
func (c *Config) setDefaults() {
	if c.Kafka.ClientID == "" {
		c.Kafka.ClientID = "kafka-pusher"
	}
	if c.Kafka.Timeout == 0 {
		c.Kafka.Timeout = 10 * time.Second
	}
	if c.Kafka.BatchSize == 0 {
		c.Kafka.BatchSize = 100
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Payload.BatchSize == 0 {
		c.Payload.BatchSize = 1
	}
	if c.Scheduler != nil && c.Scheduler.Enabled {
		if c.Scheduler.Interval == 0 {
			c.Scheduler.Interval = 5 * time.Second
		}
		if c.Scheduler.WorkerPoolSize == 0 {
			c.Scheduler.WorkerPoolSize = 1
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if c.Kafka.Topic == "" {
		return fmt.Errorf("kafka.topic is required")
	}
	if c.Payload.TemplatePath == "" {
		return fmt.Errorf("payload.template_path is required")
	}
	if c.Scheduler != nil && c.Scheduler.Enabled {
		if c.Scheduler.Interval <= 0 {
			return fmt.Errorf("scheduler.interval must be positive")
		}
		if c.Scheduler.WorkerPoolSize < 1 {
			return fmt.Errorf("scheduler.worker_pool_size must be at least 1")
		}
	}
	return nil
}
