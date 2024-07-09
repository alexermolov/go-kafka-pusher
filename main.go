package main

import (
	"github.com/alexermolov/go-kafka-pusher/pkg/loader"
	"github.com/alexermolov/go-kafka-pusher/pkg/processor"
)

func main() {
	settings := loader.GetSettings()

	processor := processor.NewProcessor(settings)
	processor.Push()
}
