package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
kafka:
  brokers:
    - localhost:9092
  topic: test-topic
  client_id: test-client
  timeout: 5s

logging:
  level: debug
  format: json

payload:
  template_path: ./test.yaml
`
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Load the config
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate loaded values
	if len(cfg.Kafka.Brokers) != 1 || cfg.Kafka.Brokers[0] != "localhost:9092" {
		t.Errorf("Expected broker localhost:9092, got %v", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.Topic != "test-topic" {
		t.Errorf("Expected topic test-topic, got %s", cfg.Kafka.Topic)
	}
	if cfg.Kafka.ClientID != "test-client" {
		t.Errorf("Expected client_id test-client, got %s", cfg.Kafka.ClientID)
	}
	if cfg.Kafka.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", cfg.Kafka.Timeout)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level debug, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format json, got %s", cfg.Logging.Format)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
				},
				Payload: PayloadConfig{
					TemplatePath: "./test.yaml",
				},
			},
			wantErr: false,
		},
		{
			name: "missing brokers",
			cfg: Config{
				Kafka: KafkaConfig{
					Topic: "test-topic",
				},
				Payload: PayloadConfig{
					TemplatePath: "./test.yaml",
				},
			},
			wantErr: true,
		},
		{
			name: "missing topic",
			cfg: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
				},
				Payload: PayloadConfig{
					TemplatePath: "./test.yaml",
				},
			},
			wantErr: true,
		},
		{
			name: "missing template path",
			cfg: Config{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
					Topic:   "test-topic",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := Config{
		Kafka: KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "test-topic",
		},
		Payload: PayloadConfig{
			TemplatePath: "./test.yaml",
		},
	}

	cfg.setDefaults()

	if cfg.Kafka.ClientID != "kafka-pusher" {
		t.Errorf("Expected default client_id kafka-pusher, got %s", cfg.Kafka.ClientID)
	}
	if cfg.Kafka.Timeout != 10*time.Second {
		t.Errorf("Expected default timeout 10s, got %v", cfg.Kafka.Timeout)
	}
	if cfg.Payload.BatchSize != 1 {
		t.Errorf("Expected default payload batch_size 1, got %d", cfg.Payload.BatchSize)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected default log format text, got %s", cfg.Logging.Format)
	}
}
