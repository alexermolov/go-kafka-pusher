package template

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	tmpl "text/template"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Template represents a message template with substitutions
type Template struct {
	Substitution map[string]interface{} `yaml:"substitution" json:"substitution"`
	Template     map[string]interface{} `yaml:"template" json:"template"`
	
	compiledTemplate *tmpl.Template
	mu               sync.RWMutex
}

// Generator is a thread-safe template generator
type Generator struct {
	template *Template
	mu       sync.RWMutex
}

// NewGenerator creates a new template generator from a file
// Supports both YAML and JSON formats based on file extension
func NewGenerator(path string) (*Generator, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	var t Template
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("failed to parse JSON template: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("failed to parse YAML template: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &t); err != nil {
			if jsonErr := json.Unmarshal(data, &t); jsonErr != nil {
				return nil, fmt.Errorf("failed to parse template as YAML or JSON: YAML error: %w, JSON error: %v", err, jsonErr)
			}
		}
	}

	return &Generator{
		template: &t,
	}, nil
}

// Generate creates a new message from the template
// This method is thread-safe
func (g *Generator) Generate() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Build substitution map with generated values
	substitutions, err := g.buildSubstitutions()
	if err != nil {
		return nil, fmt.Errorf("failed to build substitutions: %w", err)
	}

	// Convert template to JSON
	templateJSON, err := json.Marshal(g.template.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template: %w", err)
	}

	// Apply substitutions
	result, err := g.applySubstitutions(templateJSON, substitutions)
	if err != nil {
		return nil, fmt.Errorf("failed to apply substitutions: %w", err)
	}

	return result, nil
}

// buildSubstitutions generates all substitution values
func (g *Generator) buildSubstitutions() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range g.template.Substitution {
		strValue, ok := value.(string)
		if !ok {
			result[key] = value
			continue
		}

		// Process template functions
		processedValue, err := g.processValue(strValue)
		if err != nil {
			return nil, fmt.Errorf("failed to process key %s: %w", key, err)
		}
		result[key] = processedValue
	}

	return result, nil
}

// processValue processes a single substitution value with template functions
func (g *Generator) processValue(value string) (interface{}, error) {
	// GUID generator
	if matched, _ := regexp.MatchString(`{{\s*@guid\s*}}`, value); matched {
		return generateGUID()
	}

	// UUID generator
	if matched, _ := regexp.MatchString(`{{\s*@uuid\s*}}`, value); matched {
		return uuid.New().String(), nil
	}

	// Now/timestamp generator
	if re := regexp.MustCompile(`{{\s*@now\|?([a-zA-Z0-9]*)\s*}}`); re.MatchString(value) {
		matches := re.FindStringSubmatch(value)
		format := "RFC3339"
		if len(matches) > 1 && matches[1] != "" {
			format = matches[1]
		}
		return formatTime(time.Now(), format)
	}

	// Random number generator
	if re := regexp.MustCompile(`{{\s*@rnd\|?(\d*)\s*}}`); re.MatchString(value) {
		matches := re.FindStringSubmatch(value)
		digits := 6 // default
		if len(matches) > 1 && matches[1] != "" {
			digits, _ = strconv.Atoi(matches[1])
		}
		return generateRandomNumber(digits)
	}

	// If no special pattern, return as is
	return value, nil
}

// applySubstitutions applies the substitution map to the template
func (g *Generator) applySubstitutions(templateJSON []byte, substitutions map[string]interface{}) ([]byte, error) {
	t, err := tmpl.New("message").Parse(string(templateJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, substitutions); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateGUID generates a cryptographically secure GUID
func generateGUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// generateRandomNumber generates a random number with specified digits
// Uses crypto/rand for thread-safety and security
func generateRandomNumber(digits int) (string, error) {
	if digits <= 0 {
		return "0", nil
	}
	if digits > 18 {
		digits = 18 // Prevent overflow
	}

	// Calculate max value: 10^digits - 1
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n), nil
}

// formatTime formats time according to the specified format
func formatTime(t time.Time, format string) (string, error) {
	format = strings.ToUpper(format)
	
	switch format {
	case "RFC822":
		return t.Format(time.RFC822), nil
	case "RFC822Z":
		return t.Format(time.RFC822Z), nil
	case "RFC850":
		return t.Format(time.RFC850), nil
	case "RFC1123":
		return t.Format(time.RFC1123), nil
	case "RFC1123Z":
		return t.Format(time.RFC1123Z), nil
	case "RFC3339":
		return t.Format(time.RFC3339), nil
	case "RFC3339NANO":
		return t.Format(time.RFC3339Nano), nil
	case "UNIX":
		return strconv.FormatInt(t.Unix(), 10), nil
	case "UNIXMILLI":
		return strconv.FormatInt(t.UnixMilli(), 10), nil
	case "UNIXNANO":
		return strconv.FormatInt(t.UnixNano(), 10), nil
	case "ANSIC":
		return t.Format(time.ANSIC), nil
	case "UNIXDATE":
		return t.Format(time.UnixDate), nil
	case "RUBYDATE":
		return t.Format(time.RubyDate), nil
	default:
		// Try as custom format
		return t.Format(format), nil
	}
}

// GenerateHex generates a random hex string of specified length
func GenerateHex(length int) (string, error) {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}
