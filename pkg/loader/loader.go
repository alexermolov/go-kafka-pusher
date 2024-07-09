package loader

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"text/template"
	"time"

	"github.com/beevik/guid"
)

type Loader struct {
}

type Payload struct {
	Substitution map[string]interface{} `json:"substitution"`
	Template     map[string]interface{} `json:"template"`
}

type Settings struct {
	BootstrapServers string `json:"bootstrapServers"`
	Topic            string `json:"topic"`
	ClientId         string `json:"clientId"`
	Partition        int    `json:"partition"`
}

type Message struct {
	Settings *Settings
	Message  bytes.Buffer
}

func GetSettings() *Message {
	config := flag.String("config", "./config.json", "settings file path")
	payload := flag.String("payload", "./payload.json", "Kafka payload template")

	flag.Parse()

	loader := &Loader{}
	parsed := loader.ParseSettings(*config, *payload)

	return parsed
}

func (l *Loader) ParseSettings(config string, payload string) *Message {
	// config
	settings := l.LoadSettings(config)

	// payload
	load := l.LoadPayload(payload)

	var result bytes.Buffer
	resultSub := l.BuildSub(load.Substitution)

	str, _ := json.Marshal(load.Template)
	t := template.Must(template.New("").Parse(string(str[:])))

	if err := t.Execute(&result, resultSub); err != nil {
		panic(err)
	}

	return &Message{
		Settings: &settings,
		Message:  result,
	}
}

func (l *Loader) LoadSettings(config string) Settings {
	var settings Settings
	configFile, err := os.Open(config)
	if err != nil {
		log.Fatalln("❌ Cannot open settings file")
	}
	defer configFile.Close()

	byteValue, _ := io.ReadAll(configFile)
	json.Unmarshal(byteValue, &settings)

	return settings
}

func (l *Loader) LoadPayload(payload string) Payload {
	var load Payload
	payloadFile, err := os.Open(payload)
	if err != nil {
		log.Fatalln("❌ Cannot open settings file")
	}
	defer payloadFile.Close()

	payl, _ := io.ReadAll(payloadFile)
	json.Unmarshal(payl, &load)

	return load
}

func (l *Loader) BuildSub(sub map[string]interface{}) map[string]interface{} {
	for key, val := range sub {
		if val.(string) == "{{@guid}}" {
			sub[key] = guid.New().String()
		}

		if val.(string) == "{{@now}}" {
			sub[key] = time.Now().Format(time.RFC3339)
		}
	}

	return sub
}
