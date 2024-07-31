package scheduler

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
)

type Runner struct {
	lock         sync.RWMutex
	jobsQueue    *llq.Queue
	executingCnt int32
	maxJobsLimit int32
}

func CreateRunner() *Runner {
	runner := &Runner{
		lock:         sync.RWMutex{},
		jobsQueue:    llq.New(),
		maxJobsLimit: 100,
	}
	return runner
}

func (r *Runner) RunContinuously() {
	for {
		if r.jobsQueue.Size() > 0 && r.executingCnt < r.maxJobsLimit {

			job, ok := r.Pop()
			_ = ok

			switch job := job.(type) {
			case *Job:
				atomic.AddInt32(&r.executingCnt, 1)

				go r.callJobFuncWithParams(job)

				now := time.Now()
				job.ExecutedAt = now.Unix()

				if job.CronSchedule != nil {
					job.NextExecutionAt = job.CronSchedule.Next(now).Unix()
				} else {
					job.NextExecutionAt = now.Add(time.Second * time.Duration(job.Period)).Unix()
				}

				job.ExecutedTimes++
			}
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (r *Runner) Push(job *Job) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.jobsQueue.Enqueue(job)
}

func (r *Runner) Pop() (interface{}, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.jobsQueue.Dequeue()
}

func (r *Runner) callJobFuncWithParams(job *Job) error {
	t := time.Now()

	if job.JobFunction == nil {
		return nil
	}
	f := reflect.ValueOf(job.JobFunction)
	if f.IsZero() {
		return nil
	}
	if len(job.JobParams) != f.Type().NumIn() {
		return nil
	}
	in := make([]reflect.Value, len(job.JobParams))
	for k, param := range job.JobParams {
		in[k] = reflect.ValueOf(param)
	}
	vals := f.Call(in)
	for _, val := range vals {
		i := val.Interface()
		if err, ok := i.(error); ok {
			return err
		}
	}

	atomic.AddInt32(&r.executingCnt, -1)

	tm := time.Since(t)
	job.Measurement.Enqueue(tm.Nanoseconds())
	job.LastCall = fmt.Sprintf("%v", tm)

	return nil
}
