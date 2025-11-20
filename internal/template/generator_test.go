package template

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
	}{
		{
			name: "YAML format",
			content: `
substitution:
  test: "value"

template:
  message: "{{.test}}"
`,
			filename: "template-*.yaml",
		},
		{
			name: "JSON format",
			content: `{
  "substitution": {
    "test": "value"
  },
  "template": {
    "message": "{{.test}}"
  }
}`,
			filename: "template-*.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Create generator
			gen, err := NewGenerator(tmpfile.Name())
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}

			if gen == nil {
				t.Fatal("Expected non-nil generator")
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filename string
	}{
		{
			name: "YAML template",
			content: `
substitution:
  guid: "{{@guid}}"
  uuid: "{{@uuid}}"
  now: "{{@now|RFC3339}}"
  random: "{{@rnd|5}}"

template:
  id: "{{.guid}}"
  uuid: "{{.uuid}}"
  timestamp: "{{.now}}"
  number: "{{.random}}"
`,
			filename: "template-*.yaml",
		},
		{
			name: "JSON template",
			content: `{
  "substitution": {
    "guid": "{{@guid}}",
    "uuid": "{{@uuid}}",
    "now": "{{@now|RFC3339}}",
    "random": "{{@rnd|5}}"
  },
  "template": {
    "id": "{{.guid}}",
    "uuid": "{{.uuid}}",
    "timestamp": "{{.now}}",
    "number": "{{.random}}"
  }
}`,
			filename: "template-*.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			gen, err := NewGenerator(tmpfile.Name())
			if err != nil {
				t.Fatal(err)
			}

			// Generate message
			msg, err := gen.Generate()
			if err != nil {
				t.Fatalf("Failed to generate message: %v", err)
			}

			// Parse the generated JSON
			var result map[string]interface{}
			if err := json.Unmarshal(msg, &result); err != nil {
				t.Fatalf("Failed to unmarshal generated message: %v", err)
			}

			// Validate GUID format
			if id, ok := result["id"].(string); !ok || !isValidGUID(id) {
				t.Errorf("Expected valid GUID, got %v", result["id"])
			}

			// Validate UUID format
			if uuid, ok := result["uuid"].(string); !ok || !isValidUUID(uuid) {
				t.Errorf("Expected valid UUID, got %v", result["uuid"])
			}

			// Validate timestamp exists
			if _, ok := result["timestamp"].(string); !ok {
				t.Errorf("Expected timestamp string, got %v", result["timestamp"])
			}

			// Validate random number
			if num, ok := result["number"].(string); !ok || len(num) != 5 {
				t.Errorf("Expected 5-digit number, got %v", result["number"])
			}
		})
	}
}

func TestGenerateConcurrent(t *testing.T) {
	content := `
substitution:
  guid: "{{@guid}}"
  random: "{{@rnd|10}}"

template:
  id: "{{.guid}}"
  value: "{{.random}}"
`
	tmpfile, err := os.CreateTemp("", "template-*.yaml")
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

	gen, err := NewGenerator(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Test concurrent generation
	const goroutines = 100
	done := make(chan bool, goroutines)
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := gen.Generate()
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent generation failed: %v", err)
	}
}

func TestGenerateRandomNumber(t *testing.T) {
	tests := []struct {
		digits  int
		wantLen int
	}{
		{1, 1},
		{5, 5},
		{10, 10},
		{15, 15},
	}

	for _, tt := range tests {
		t.Run("digits", func(t *testing.T) {
			result, err := generateRandomNumber(tt.digits)
			if err != nil {
				t.Fatalf("generateRandomNumber() error = %v", err)
			}
			if len(result) != tt.wantLen {
				t.Errorf("Expected length %d, got %d (value: %s)", tt.wantLen, len(result), result)
			}
		})
	}
}

func isValidGUID(s string) bool {
	match, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, s)
	return match
}

func isValidUUID(s string) bool {
	match, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, s)
	return match
}
