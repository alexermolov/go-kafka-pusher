# Kafka Pusher

[![Go Report Card](https://goreportcard.com/badge/github.com/alexermolov/go-kafka-pusher)](https://goreportcard.com/report/github.com/alexermolov/go-kafka-pusher)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, thread-safe Kafka message producer with template-based payload generation and flexible scheduling.

## Features

- 🚀 **High Performance**: Optimized Kafka writer with connection pooling and batching
- 🔒 **Thread-Safe**: All components are designed for concurrent use
- 📝 **Template Engine**: Dynamic message generation with substitution variables
- ⏰ **Flexible Scheduling**: Built-in scheduler with configurable worker pools
- 🔧 **Configuration**: YAML-based configuration with environment variable support
- 📊 **Structured Logging**: Built-in structured logging with slog
- 🛡️ **Graceful Shutdown**: Proper signal handling and resource cleanup
- 🎯 **Production Ready**: Comprehensive error handling and statistics tracking

## Quick Start

### Installation

```bash
go install github.com/alexermolov/go-kafka-pusher/cmd/kafka-pusher@latest
```

Or build from source:

```bash
git clone https://github.com/alexermolov/go-kafka-pusher.git
cd go-kafka-pusher
make build
```

### Basic Usage

1. Create configuration files:

```bash
cp config.example.yaml config.yaml
cp payload.example.yaml payload.yaml
```

2. Edit `config.yaml` with your Kafka settings:

```yaml
kafka:
  brokers:
    - localhost:9092
  topic: my-topic
  client_id: kafka-pusher
```

3. Run the application:

```bash
# Single message
./bin/kafka-pusher -config config.yaml

# Continuous mode with scheduler
# (Enable scheduler in config.yaml first)
./bin/kafka-pusher -config config.yaml
```

## Configuration

### Main Configuration (`config.yaml`)

```yaml
kafka:
  brokers:
    - ${KAFKA_BROKERS:-localhost:9092}
  topic: ${KAFKA_TOPIC:-test-topic}
  client_id: kafka-pusher
  partition: 0              # -1 for automatic
  timeout: 10s
  batch_size: 100
  async: false

scheduler:
  enabled: true
  interval: 5s
  worker_pool_size: 1

logging:
  level: info               # debug, info, warn, error
  format: text              # text or json
  verbose: true

payload:
  template_path: ./payload.yaml  # or ./payload.json
```

### Payload Template (`payload.yaml` or `payload.json`)

The payload template supports both YAML and JSON formats. The format is automatically detected by file extension.

**YAML format:**
```yaml
substitution:
  guid: "{{@guid}}"                    # Random GUID
  uuid: "{{@uuid}}"                    # Random UUID v4
  now: "{{@now|RFC3339}}"              # Current timestamp
  timestamp: "{{@now|UnixMilli}}"      # Unix milliseconds
  randomNum: "{{@rnd|10}}"             # Random 10-digit number
  
template:
  messageId: "{{.guid}}"
  timestamp: "{{.now}}"
  data:
    value: "{{.randomNum}}"
```

**JSON format:**
```json
{
  "substitution": {
    "guid": "{{@guid}}",
    "uuid": "{{@uuid}}",
    "now": "{{@now|RFC3339}}",
    "timestamp": "{{@now|UnixMilli}}",
    "randomNum": "{{@rnd|10}}"
  },
  "template": {
    "messageId": "{{.guid}}",
    "timestamp": "{{.now}}",
    "data": {
      "value": "{{.randomNum}}"
    }
  }
}
```

### Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `{{@guid}}` | Generates a GUID | `550e8400-e29b-41d4-a716-446655440000` |
| `{{@uuid}}` | Generates UUID v4 | `f47ac10b-58cc-4372-a567-0e02b2c3d479` |
| `{{@now\|FORMAT}}` | Current timestamp | `{{@now\|RFC3339}}` |
| `{{@rnd\|DIGITS}}` | Random number | `{{@rnd\|6}}` → `123456` |

#### Supported Time Formats

- `RFC3339`, `RFC3339Nano`
- `RFC822`, `RFC822Z`
- `RFC850`, `RFC1123`, `RFC1123Z`
- `Unix`, `UnixMilli`, `UnixNano`
- `ANSIC`, `UnixDate`, `RubyDate`

## Development

### Prerequisites

- Go 1.22 or higher
- Docker & Docker Compose (for local Kafka)

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/alexermolov/go-kafka-pusher.git
cd go-kafka-pusher

# Install dependencies
go mod download

# Start local Kafka cluster
make kafka-up

# Run tests
make test

# Run with race detector
make test-race

# Build
make build
```

### Project Structure

```
.
├── cmd/
│   └── kafka-pusher/       # Main application
├── internal/
│   ├── config/             # Configuration management
│   ├── kafka/              # Kafka producer
│   ├── logger/             # Structured logging
│   ├── scheduler/          # Task scheduler
│   └── template/           # Template generator
├── config.example.yaml     # Example configuration
├── payload.example.yaml    # Example payload template
├── docker-compose.yaml     # Local Kafka setup
└── Makefile               # Build automation
```

## Docker

### Using Docker Compose

Run the complete stack (Kafka + Zookeeper + Pusher):

```bash
docker-compose up
```

### Build Docker Image

```bash
docker build -t kafka-pusher .
```

### Run Container

```bash
docker run -v $(pwd)/config.yaml:/app/config.yaml \
           -v $(pwd)/payload.yaml:/app/payload.yaml \
           kafka-pusher
```

## Environment Variables

All configuration values can be overridden with environment variables:

```bash
export KAFKA_BROKERS="kafka1:9092,kafka2:9092"
export KAFKA_TOPIC="my-topic"
./bin/kafka-pusher
```

## Performance Tuning

### High Throughput

```yaml
kafka:
  batch_size: 1000
  async: true
  
scheduler:
  worker_pool_size: 10
```

### Low Latency

```yaml
kafka:
  batch_size: 1
  async: false
  timeout: 1s
```

## Monitoring

The application logs structured metrics:

```json
{
  "level": "info",
  "msg": "message sent successfully",
  "topic": "test-topic",
  "partition": 0,
  "size": 1024,
  "duration": "5ms"
}
```

## Troubleshooting

### Connection Issues

```bash
# Test Kafka connectivity
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

### Enable Debug Logging

```yaml
logging:
  level: debug
  verbose: true
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [segmentio/kafka-go](https://github.com/segmentio/kafka-go) - Pure Go Kafka client
- [google/uuid](https://github.com/google/uuid) - UUID generation

## Support

- 📫 Issues: [GitHub Issues](https://github.com/alexermolov/go-kafka-pusher/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/alexermolov/go-kafka-pusher/discussions)
