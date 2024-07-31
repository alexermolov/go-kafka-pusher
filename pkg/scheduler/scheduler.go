package scheduler

import (
	"time"

	cb "github.com/emirpasic/gods/queues/circularbuffer"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	jobs          map[uuid.UUID]*Job `json:"jobs"`
	fluentTracker *uuid.UUID
	runner        *Runner
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		jobs:          map[uuid.UUID]*Job{},
		fluentTracker: nil,
		runner:        CreateRunner(),
	}

	return s
}

func (s *Scheduler) Name(name string) *Scheduler {
	job := s.getCurrentJob()
	job.Name = name

	return s
}

func (s *Scheduler) Do(jobFun interface{}, params ...interface{}) *Job {
	job := s.getCurrentJob()
	job.JobFunction = jobFun
	s.fluentTracker = nil

	return job
}

func (s *Scheduler) Every(interval int32) *Scheduler {
	job := s.getCurrentJob()
	job.Period = interval

	return s
}

func (s *Scheduler) Crone(interval string, withSeconds bool) *Scheduler {
	job := s.getCurrentJob()

	var (
		cronSchedule cron.Schedule
		err          error
	)

	if withSeconds {
		p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Month)
		cronSchedule, err = p.Parse(interval)
	} else {
		cronSchedule, err = cron.ParseStandard(interval)
	}

	if err != nil {
		job.Err = err
	}

	job.CronSchedule = cronSchedule

	now := time.Now()
	job.ExecutedAt = now.Unix()
	job.NextExecutionAt = job.CronSchedule.Next(now).Unix()
	job.CronPeriod = interval

	return s
}

func (s *Scheduler) getCurrentJob() *Job {
	if s.fluentTracker == nil {
		job := Job{
			Id:          uuid.New(),
			ExecutedAt:  0,
			Measurement: *cb.New(64),
		}

		s.jobs[job.Id] = &job

		s.fluentTracker = &job.Id
		return &job
	}

	return s.jobs[*s.fluentTracker]
}

func (s *Scheduler) Run() {

	ticker := time.NewTicker(500 * time.Millisecond)

	go s.runner.RunContinuously()

	go func() {
		for {
			select {
			case t := <-ticker.C:
				_ = t
				for _, job := range s.jobs {
					if job.NextExecutionAt <= time.Now().Unix() {
						s.runner.Push(job)
					}
				}
			}
		}
	}()
}

func (s *Scheduler) GetJobsCount() int {
	return len(s.jobs)
}

func (s *Scheduler) GetJobs() map[uuid.UUID]*Job {
	return s.jobs
}
