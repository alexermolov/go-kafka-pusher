package loader

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"math/rand"
	"os"
	"regexp"
	"strconv"
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
	BootstrapServers string     `json:"bootstrapServers"`
	Topic            string     `json:"topic"`
	ClientId         string     `json:"clientId"`
	Partition        int        `json:"partition"`
	Verbose          bool       `json:"verbose"`
	Scheduler        *Scheduler `json:"scheduler"`
}

type Scheduler struct {
	Enabled   bool  `json:"enabled"`
	VUs       int   `json:"VUs"`
	PeriodSec int32 `json:"periodSec"`
}

type Callable func() *bytes.Buffer

type Message struct {
	Settings *Settings
	Message  Callable
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

	fn := func() *bytes.Buffer {
		str, _ := json.Marshal(load.Template)
		clone := maps.Clone(load.Substitution)
		return l.BuildTemplate(str, clone)
	}

	return &Message{
		Settings: &settings,
		Message:  fn,
	}
}

func (l *Loader) BuildTemplate(str []byte, sub map[string]interface{}) *bytes.Buffer {
	resultSub := l.BuildSub(sub)

	var result bytes.Buffer

	t := template.Must(template.New("").Parse(string(str[:])))
	if err := t.Execute(&result, resultSub); err != nil {
		panic(err)
	}

	return &result
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
	rGuid, _ := regexp.Compile(`{{ *@guid *}}`)
	rNow, _ := regexp.Compile(`{{ *@now\|?([a-zA-Z0-9]+) *}}`)
	rRnd, _ := regexp.Compile(`{{ *@rnd\|?(\d*) *}}`)
	fxrRnd, _ := regexp.Compile(`{{ *@fixrnd\|?(\d*) *}}`)

	for key, val := range sub {
		makeGuid(rGuid, val, sub, key)
		makeNow(rNow, val, sub, key)
		makeRandom(rRnd, val, sub, key)
		makeFixedRandom(fxrRnd, val, sub, key)
	}

	return sub
}

func makeGuid(rGuid *regexp.Regexp, val interface{}, sub map[string]interface{}, key string) {
	matchGuid := rGuid.MatchString(val.(string))
	if matchGuid {
		sub[key] = guid.New().String()
	}
}

func makeNow(rNow *regexp.Regexp, val interface{}, sub map[string]interface{}, key string) {
	matchNow := rNow.MatchString(val.(string))
	if matchNow {
		m := rNow.FindStringSubmatch(val.(string))
		constSpec := time.RFC3339

		if len(m) == 2 {
			switch m[1] {
			case "RFC822":
				constSpec = time.RFC822
			case "RFC822Z":
				constSpec = time.RFC822Z
			case "RFC850":
				constSpec = time.RFC850
			case "RFC1123":
				constSpec = time.RFC1123
			case "RFC1123Z":
				constSpec = time.RFC1123Z
			case "RFC3339":
				constSpec = time.RFC3339
			case "RFC3339Nano":
				constSpec = time.RFC3339Nano
			}
		}
		sub[key] = time.Now().Format(constSpec)
	}
}

func makeFixedRandom(fxrRnd *regexp.Regexp, val interface{}, sub map[string]interface{}, key string) {
	matchFixrnd := fxrRnd.MatchString(val.(string))
	if matchFixrnd {
		m := fxrRnd.FindStringSubmatch(val.(string))

		if len(m) == 2 {
			n, _ := strconv.Atoi(m[1])
			f := fmt.Sprintf(`%%0%dd`, n)
			mul := fmt.Sprintf(fmt.Sprintf(`1%%0%dd`, n), 0)
			n2, _ := strconv.Atoi(mul)

			sub[key] = fmt.Sprintf(f, rand.Intn(n2-1))
		}
	}
}

func makeRandom(rRnd *regexp.Regexp, val interface{}, sub map[string]interface{}, key string) {
	matchRnd := rRnd.MatchString(val.(string))
	if matchRnd {
		m := rRnd.FindStringSubmatch(val.(string))

		if len(m) == 2 {
			n, _ := strconv.Atoi(m[1])
			mul := fmt.Sprintf(fmt.Sprintf(`1%%0%dd`, n), 0)
			n2, _ := strconv.Atoi(mul)

			sub[key] = rand.Intn(n2 - 1)
		}
	}
}
