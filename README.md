# Kafka Pusher

[![Go Report Card](https://goreportcard.com/badge/github.com/alexermolov/go-kafka-pusher)](https://goreportcard.com/report/github.com/alexermolov/go-kafka-pusher)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, thread-safe Kafka message producer with template-based payload generation, flexible scheduling, and support for multiple concurrent payload streams.

## Features

- 🚀 **High Performance**: Optimized Kafka writer with parallel message generation
- 🔀 **Multiple Payloads**: Configure and send different message types simultaneously
- 📦 **Batch Processing**: Control batch sizes per payload for optimal throughput
- 🔒 **Thread-Safe**: All components are designed for concurrent use
- 📝 **Template Engine**: Dynamic message generation with substitution variables
- ⏰ **Flexible Scheduling**: Built-in scheduler with configurable worker pools
- 🔧 **YAML Configuration**: Easy-to-manage YAML-based configuration
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

1. Create configuration file `config.yaml`:

```yaml
kafka:
  brokers:
    - localhost:9092
  topic: my-topic

scheduler:
  enabled: true
  interval: 5s

payloads:
  - name: my-payload
    template_path: ./payload.yaml
    batch_size: 10
```

2. Create payload template `payload.yaml`:

```yaml
substitution:
  id: "{{@uuid}}"
  timestamp: "{{@now|RFC3339}}"
  
template:
  message_id: "{{.id}}"
  created_at: "{{.timestamp}}"
  data: "Hello Kafka!"
```

3. Run the application:

```bash
# With scheduler (continuous mode)
./kafka-pusher -config config.yaml

# Single execution (disable scheduler first)
./kafka-pusher -config config.yaml
```

## Use Cases

### Load Testing

Generate high-volume traffic to test Kafka cluster performance:

```yaml
scheduler:
  interval: 1s
  worker_pool_size: 10

payloads:
  - name: load-test
    template_path: ./payload.yaml
    batch_size: 1000  # 1000 messages per second
```

### Multi-Stream Data Generation

Simulate multiple data sources sending different message types:

```yaml
payloads:
  - name: user-events
    template_path: ./user-events.yaml
    batch_size: 50
    
  - name: system-logs
    template_path: ./system-logs.yaml
    batch_size: 100
    
  - name: transactions
    template_path: ./transactions.yaml
    batch_size: 20
```

All payloads are sent **in parallel** every scheduler interval.

### Development & Testing

Populate test topics with realistic data:

```yaml
scheduler:
  enabled: false  # Single execution

payloads:
  - name: test-data
    template_path: ./test-payload.yaml
    batch_size: 100
```

## Configuration

### Main Configuration (`config.yaml`)

```yaml
kafka:
  brokers:
    - localhost:9092
  topic: test-topic
  client_id: kafka-pusher
  partition: 0              # -1 for automatic
  timeout: 10s
  async: false

scheduler:
  enabled: true
  interval: 5s              # How often to send messages
  worker_pool_size: 1       # Number of concurrent workers

logging:
  level: info               # debug, info, warn, error
  format: text              # text or json
  verbose: true             # Log full message content

# Multiple payload configurations
payloads:
  - name: orders            # Identifier for this payload
    template_path: ./payload.yaml
    batch_size: 10          # Send 10 messages per execution
  
  - name: events
    template_path: ./payload-events.yaml
    batch_size: 5           # Send 5 messages per execution
```

**Key Configuration Options:**

- **`payloads`**: Array of payload configurations, each with its own template and batch size
- **`batch_size`**: Number of messages to generate and send in each batch for this payload
- **`name`**: Identifier used in logs to distinguish between different payloads
- All payloads are processed **in parallel** for maximum throughput

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

## Performance Tuning

### High Throughput

For maximum message volume, increase batch sizes and enable multiple payloads:

```yaml
kafka:
  async: true
  
scheduler:
  interval: 1s
  worker_pool_size: 5

payloads:
  - name: stream-1
    template_path: ./payload-1.yaml
    batch_size: 100
  - name: stream-2
    template_path: ./payload-2.yaml
    batch_size: 100
  - name: stream-3
    template_path: ./payload-3.yaml
    batch_size: 100
```

This configuration will send **300 messages per second** (3 payloads × 100 messages each).

### Low Latency

For minimal delay between message generation and delivery:

```yaml
kafka:
  async: false
  timeout: 1s

scheduler:
  interval: 100ms
  worker_pool_size: 1

payloads:
  - name: realtime
    template_path: ./payload.yaml
    batch_size: 1
```

### Balanced Configuration

```yaml
scheduler:
  interval: 5s
  worker_pool_size: 2

payloads:
  - name: orders
    template_path: ./payload-orders.yaml
    batch_size: 20
  - name: events
    template_path: ./payload-events.yaml
    batch_size: 10
```

Sends **30 messages every 5 seconds** (20 orders + 10 events in parallel).

## Monitoring

The application provides detailed structured logging for each payload:

```
time=2025-12-04T18:00:00.000+00:00 level=INFO msg="template generator initialized" name=orders path=./payload.yaml batch_size=10
time=2025-12-04T18:00:00.000+00:00 level=INFO msg="template generator initialized" name=events path=./payload-events.yaml batch_size=5
time=2025-12-04T18:00:00.000+00:00 level=INFO msg="kafka producer initialized" brokers=[localhost:9092] topic=test-topic
time=2025-12-04T18:00:00.000+00:00 level=INFO msg="scheduler started" interval=5s workers=1

time=2025-12-04T18:00:00.100+00:00 level=INFO msg="sending batch to Kafka" payload=orders batch_size=10
time=2025-12-04T18:00:00.105+00:00 level=INFO msg="batch sent successfully" topic=test-topic count=10 duration=5ms

time=2025-12-04T18:00:00.110+00:00 level=INFO msg="sending batch to Kafka" payload=events batch_size=5
time=2025-12-04T18:00:00.112+00:00 level=INFO msg="batch sent successfully" topic=test-topic count=5 duration=2ms
```

### Key Metrics

- **Batch size**: Number of messages in each batch per payload
- **Duration**: Time taken to send the batch
- **Count**: Actual number of messages sent
- **Payload name**: Which payload stream generated the messages

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
