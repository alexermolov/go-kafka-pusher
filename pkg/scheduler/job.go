package scheduler

import (
	"fmt"
	"strconv"
	"time"

	cb "github.com/emirpasic/gods/queues/circularbuffer"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Job struct {
	Id              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	JobFunction     interface{}   `json:"-"`
	JobParams       []interface{} `json:"-"`
	Period          int32         `json:"period"`
	CronPeriod      string        `json:"cronPeriod"`
	ExecutedAt      int64         `json:"executedAt"`
	NextExecutionAt int64         `json:"nextExecutionAt"`
	ExecutedTimes   int64         `json:"executedTimes"`
	Measurement     cb.Queue      `json:"measurement"`
	LastCall        string        `json:"lastCall"`
	CronSchedule    cron.Schedule `json:"-"`
	Err             error         `json:"-"`
}

func (j *Job) GetExecutedTimes() int64 {
	return j.ExecutedTimes
}

func (j *Job) GetName() string {
	return j.Name
}

func (j *Job) NextExecution() time.Time {
	i, _ := strconv.ParseInt(fmt.Sprint(j.NextExecutionAt), 10, 64)
	tm := time.Unix(i, 0)

	return tm
}

func (j *Job) PrevExecution() time.Time {
	i, _ := strconv.ParseInt(fmt.Sprint(j.ExecutedAt), 10, 64)
	tm := time.Unix(i, 0)

	return tm
}
