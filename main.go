package main

import (
	"fmt"

	"github.com/alexermolov/go-kafka-pusher/pkg/loader"
	"github.com/alexermolov/go-kafka-pusher/pkg/processor"
	"github.com/alexermolov/go-kafka-pusher/pkg/scheduler"
)

func main() {
	settings := loader.GetSettings()
	processor := processor.NewProcessor(settings)

	if settings.Settings.Scheduler != nil && settings.Settings.Scheduler.Enabled {
		s := scheduler.NewScheduler()

		s.Name("Periodical Pusher").Every(settings.Settings.Scheduler.PeriodSec).Do(func() {
			processor.Push()
		})

		s.Run()

		fmt.Scanln()
	} else {
		processor.Push()
	}
}
