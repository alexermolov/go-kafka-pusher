package template

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestProjectPayloadYAML tests parsing the actual payload.yaml from project root
func TestProjectPayloadYAML(t *testing.T) {
	projectRoot := filepath.Join("..", "..")
	payloadPath := filepath.Join(projectRoot, "payload.yaml")
	
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		t.Skip("payload.yaml not found in project root, skipping test")
	}
	
	gen, err := NewGenerator(payloadPath)
	if err != nil {
		t.Fatalf("Failed to create generator from payload.yaml: %v", err)
	}
	
	// Generate a message
	msg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Failed to generate message from payload.yaml: %v", err)
	}
	
	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(msg, &result); err != nil {
		t.Fatalf("Generated message is not valid JSON: %v\nMessage: %s", err, string(msg))
	}
	
	t.Logf("Successfully generated message from payload.yaml:")
	t.Logf("Message: %s", string(msg))
	
	// Verify expected fields
	if _, ok := result["messageId"]; !ok {
		t.Error("Expected 'messageId' field in generated message")
	}
	if _, ok := result["messageDate"]; !ok {
		t.Error("Expected 'messageDate' field in generated message")
	}
	if card, ok := result["card"].(map[string]interface{}); ok {
		if _, ok := card["number"]; !ok {
			t.Error("Expected 'card.number' field in generated message")
		}
		if _, ok := card["code"]; !ok {
			t.Error("Expected 'card.code' field in generated message")
		}
	} else {
		t.Error("Expected 'card' to be an object")
	}
}

// TestProjectPayloadJSON tests parsing the actual payload.json from project root
func TestProjectPayloadJSON(t *testing.T) {
	projectRoot := filepath.Join("..", "..")
	payloadPath := filepath.Join(projectRoot, "payload.json")
	
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		t.Skip("payload.json not found in project root, skipping test")
	}
	
	gen, err := NewGenerator(payloadPath)
	if err != nil {
		t.Fatalf("Failed to create generator from payload.json: %v", err)
	}
	
	// Generate a message
	msg, err := gen.Generate()
	if err != nil {
		t.Fatalf("Failed to generate message from payload.json: %v", err)
	}
	
	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(msg, &result); err != nil {
		t.Fatalf("Generated message is not valid JSON: %v\nMessage: %s", err, string(msg))
	}
	
	t.Logf("Successfully generated message from payload.json:")
	t.Logf("Message: %s", string(msg))
	
	// Verify expected fields
	if _, ok := result["messageId"]; !ok {
		t.Error("Expected 'messageId' field in generated message")
	}
	if _, ok := result["messageDate"]; !ok {
		t.Error("Expected 'messageDate' field in generated message")
	}
	if card, ok := result["card"].(map[string]interface{}); ok {
		if _, ok := card["number"]; !ok {
			t.Error("Expected 'card.number' field in generated message")
		}
		if _, ok := card["code"]; !ok {
			t.Error("Expected 'card.code' field in generated message")
		}
	} else {
		t.Error("Expected 'card' to be an object")
	}
}

// TestJSONPayloadFormats tests various JSON payload formats
func TestJSONPayloadFormats(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "simple_json",
			content: `{
  "substitution": {
    "id": "{{@guid}}"
  },
  "template": {
    "userId": "{{.id}}"
  }
}`,
			wantErr: false,
		},
		{
			name: "nested_json",
			content: `{
  "substitution": {
    "guid": "{{@guid}}",
    "timestamp": "{{@now|RFC3339}}"
  },
  "template": {
    "data": {
      "user": {
        "id": "{{.guid}}",
        "created": "{{.timestamp}}"
      }
    }
  }
}`,
			wantErr: false,
		},
		{
			name: "array_json",
			content: `{
  "substitution": {
    "id1": "{{@guid}}",
    "id2": "{{@guid}}"
  },
  "template": {
    "items": [
      {"id": "{{.id1}}"},
      {"id": "{{.id2}}"}
    ]
  }
}`,
			wantErr: false,
		},
		{
			name: "with_functions",
			content: `{
  "substitution": {
    "guid": "{{@guid}}",
    "uuid": "{{@uuid}}",
    "now": "{{@now|UnixMilli}}",
    "random": "{{@rnd|8}}"
  },
  "template": {
    "id": "{{.guid}}",
    "uuid": "{{.uuid}}",
    "timestamp": "{{.now}}",
    "code": "{{.random}}"
  }
}`,
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "payload-*.json")
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
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			
			// Generate message
			msg, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			
			// Verify valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal(msg, &result); err != nil {
				t.Errorf("Generated message is not valid JSON: %v\nMessage: %s", err, string(msg))
			}
			
			t.Logf("Generated: %s", string(msg))
		})
	}
}

// TestJSONvsYAML tests that JSON and YAML produce equivalent results
func TestJSONvsYAML(t *testing.T) {
	// Create YAML version
	yamlContent := `
substitution:
  staticId: "test-123"
  
template:
  id: "{{.staticId}}"
  type: "test"
`
	
	yamlFile, err := os.CreateTemp("", "payload-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(yamlFile.Name())
	
	if _, err := yamlFile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	yamlFile.Close()
	
	// Create JSON version
	jsonContent := `{
  "substitution": {
    "staticId": "test-123"
  },
  "template": {
    "id": "{{.staticId}}",
    "type": "test"
  }
}`
	
	jsonFile, err := os.CreateTemp("", "payload-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(jsonFile.Name())
	
	if _, err := jsonFile.Write([]byte(jsonContent)); err != nil {
		t.Fatal(err)
	}
	jsonFile.Close()
	
	// Generate from both
	yamlGen, err := NewGenerator(yamlFile.Name())
	if err != nil {
		t.Fatalf("Failed to create YAML generator: %v", err)
	}
	
	jsonGen, err := NewGenerator(jsonFile.Name())
	if err != nil {
		t.Fatalf("Failed to create JSON generator: %v", err)
	}
	
	yamlMsg, err := yamlGen.Generate()
	if err != nil {
		t.Fatalf("Failed to generate from YAML: %v", err)
	}
	
	jsonMsg, err := jsonGen.Generate()
	if err != nil {
		t.Fatalf("Failed to generate from JSON: %v", err)
	}
	
	// Parse both
	var yamlResult, jsonResult map[string]interface{}
	if err := json.Unmarshal(yamlMsg, &yamlResult); err != nil {
		t.Fatalf("Failed to parse YAML result: %v", err)
	}
	if err := json.Unmarshal(jsonMsg, &jsonResult); err != nil {
		t.Fatalf("Failed to parse JSON result: %v", err)
	}
	
	// Compare key fields
	if yamlResult["id"] != jsonResult["id"] {
		t.Errorf("ID mismatch: YAML=%v, JSON=%v", yamlResult["id"], jsonResult["id"])
	}
	if yamlResult["type"] != jsonResult["type"] {
		t.Errorf("Type mismatch: YAML=%v, JSON=%v", yamlResult["type"], jsonResult["type"])
	}
	
	t.Logf("YAML result: %s", string(yamlMsg))
	t.Logf("JSON result: %s", string(jsonMsg))
}
