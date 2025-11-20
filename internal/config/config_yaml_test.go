package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadProjectConfigYAML tests that the actual config.yaml from project root
// can be successfully parsed with the current Config structure
func TestLoadProjectConfigYAML(t *testing.T) {
	// Try to find config.yaml in project root (two levels up from internal/config)
	projectRoot := filepath.Join("..", "..")
	configPath := filepath.Join(projectRoot, "config.yaml")
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("config.yaml not found in project root, skipping test")
	}
	
	// Set environment variables for config.yaml (uses ${VAR:-default} syntax)
	// Note: os.ExpandEnv doesn't support :-default, so we need to set the vars
	os.Setenv("KAFKA_BROKERS", "localhost:9092")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	defer os.Unsetenv("KAFKA_BROKERS")
	defer os.Unsetenv("KAFKA_TOPIC")
	
	// Load the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config.yaml: %v", err)
	}
	
	// Validate that config was parsed correctly
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
	
	// Log parsed values for verification
	t.Logf("Successfully parsed config.yaml:")
	t.Logf("  Kafka.Brokers: %v", cfg.Kafka.Brokers)
	t.Logf("  Kafka.Topic: %s", cfg.Kafka.Topic)
	t.Logf("  Kafka.ClientID: %s", cfg.Kafka.ClientID)
	t.Logf("  Kafka.Partition: %d", cfg.Kafka.Partition)
	t.Logf("  Kafka.Timeout: %v", cfg.Kafka.Timeout)
	t.Logf("  Kafka.BatchSize: %d", cfg.Kafka.BatchSize)
	t.Logf("  Kafka.Async: %v", cfg.Kafka.Async)
	
	if cfg.Scheduler != nil {
		t.Logf("  Scheduler.Enabled: %v", cfg.Scheduler.Enabled)
		t.Logf("  Scheduler.Interval: %v", cfg.Scheduler.Interval)
		t.Logf("  Scheduler.WorkerPoolSize: %d", cfg.Scheduler.WorkerPoolSize)
	}
	
	t.Logf("  Logging.Level: %s", cfg.Logging.Level)
	t.Logf("  Logging.Format: %s", cfg.Logging.Format)
	t.Logf("  Logging.Verbose: %v", cfg.Logging.Verbose)
	t.Logf("  Payload.TemplatePath: %s", cfg.Payload.TemplatePath)
	
	// Basic assertions
	if len(cfg.Kafka.Brokers) == 0 {
		t.Error("Expected at least one Kafka broker")
	}
	if cfg.Kafka.Topic == "" {
		t.Error("Expected Kafka topic to be set")
	}
	if cfg.Payload.TemplatePath == "" {
		t.Error("Expected payload template path to be set")
	}
}

// TestConfigYAMLWithEnvVars tests that environment variable expansion works
func TestConfigYAMLWithEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config_with_env.yaml")
	
	// Set test environment variables
	os.Setenv("TEST_KAFKA_BROKERS", "testhost:9092")
	os.Setenv("TEST_KAFKA_TOPIC", "test-env-topic")
	defer os.Unsetenv("TEST_KAFKA_BROKERS")
	defer os.Unsetenv("TEST_KAFKA_TOPIC")
	
	yamlContent := `kafka:
  brokers:
    - ${TEST_KAFKA_BROKERS}
  topic: ${TEST_KAFKA_TOPIC}
  client_id: test-client
  partition: 0
  timeout: 5s
  batch_size: 50
  async: false

logging:
  level: info
  format: text
  verbose: false

payload:
  template_path: ./test.yaml
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify environment variables were expanded
	if len(cfg.Kafka.Brokers) == 0 || cfg.Kafka.Brokers[0] != "testhost:9092" {
		t.Errorf("Expected broker 'testhost:9092', got %v", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.Topic != "test-env-topic" {
		t.Errorf("Expected topic 'test-env-topic', got '%s'", cfg.Kafka.Topic)
	}
}

// TestConfigYAMLDefaults tests that default values are correctly applied
func TestConfigYAMLDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config_minimal.yaml")
	
	// Minimal config with only required fields
	yamlContent := `kafka:
  brokers:
    - localhost:9092
  topic: minimal-topic

payload:
  template_path: ./payload.yaml
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify defaults were applied
	if cfg.Kafka.ClientID != "kafka-pusher" {
		t.Errorf("Expected default ClientID 'kafka-pusher', got '%s'", cfg.Kafka.ClientID)
	}
	if cfg.Kafka.Timeout != 10*time.Second {
		t.Errorf("Expected default Timeout 10s, got %v", cfg.Kafka.Timeout)
	}
	if cfg.Kafka.BatchSize != 100 {
		t.Errorf("Expected default BatchSize 100, got %d", cfg.Kafka.BatchSize)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default logging level 'info', got '%s'", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected default logging format 'text', got '%s'", cfg.Logging.Format)
	}
}

// TestConfigYAMLWithScheduler tests scheduler configuration
func TestConfigYAMLWithScheduler(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config_scheduler.yaml")
	
	yamlContent := `kafka:
  brokers:
    - localhost:9092
  topic: scheduler-test

scheduler:
  enabled: true
  interval: 30s
  worker_pool_size: 5

logging:
  level: debug
  format: json
  verbose: true

payload:
  template_path: ./payload.yaml
`
	
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}
	
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify scheduler settings
	if cfg.Scheduler == nil {
		t.Fatal("Expected Scheduler to be non-nil")
	}
	if !cfg.Scheduler.Enabled {
		t.Error("Expected Scheduler.Enabled to be true")
	}
	if cfg.Scheduler.Interval != 30*time.Second {
		t.Errorf("Expected Scheduler.Interval 30s, got %v", cfg.Scheduler.Interval)
	}
	if cfg.Scheduler.WorkerPoolSize != 5 {
		t.Errorf("Expected Scheduler.WorkerPoolSize 5, got %d", cfg.Scheduler.WorkerPoolSize)
	}
}

// TestConfigYAMLInvalid tests that invalid config is rejected
func TestConfigYAMLInvalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing_brokers",
			content: `kafka:
  topic: test

payload:
  template_path: ./payload.yaml
`,
		},
		{
			name: "missing_topic",
			content: `kafka:
  brokers:
    - localhost:9092

payload:
  template_path: ./payload.yaml
`,
		},
		{
			name: "missing_payload",
			content: `kafka:
  brokers:
    - localhost:9092
  topic: test
`,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "invalid_config.yaml")
			
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}
			
			_, err = Load(configPath)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
		})
	}
}
