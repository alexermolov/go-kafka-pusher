package processor

import (
	"context"
	"log"
	"time"

	"github.com/alexermolov/go-kafka-pusher/pkg/loader"
	"github.com/segmentio/kafka-go"
)

type Processor struct {
	Message *loader.Message
}

func NewProcessor(settings *loader.Message) *Processor {
	return &Processor{
		Message: settings,
	}
}

func (proc *Processor) Push() {
	conn, err := kafka.DialLeader(context.Background(), "tcp", proc.Message.Settings.BootstrapServers, proc.Message.Settings.Topic, proc.Message.Settings.Partition)
	if err != nil {
		log.Fatal("❌ failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: proc.Message.Message.Bytes()},
	)
	if err != nil {
		log.Fatal("❌ failed to write messages:", err)
	}

	log.Default().Printf("✅ Message sent to %s topic %s partition %d", proc.Message.Settings.BootstrapServers, proc.Message.Settings.Topic, proc.Message.Settings.Partition)
	log.Default().Println()
	log.Default().Println()

	log.Default().Println("✅ Message was:")
	log.Default().Println(proc.Message.Message)
	log.Default().Println()
	log.Default().Println()

	if err := conn.Close(); err != nil {
		log.Fatal("❌ failed to close writer:", err)
	}
}
