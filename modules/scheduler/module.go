package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"

	log "github.com/sirupsen/logrus"
)

type Scheduler struct {
	scheduledJobs []ScheduledJob
	waitGroup     sync.WaitGroup
}

type Job interface {
	Run()
}

type ScheduledJob struct {
	Job          Job
	Tickduration time.Duration
	jobControl   jobControl
}

type jobControl struct {
	lock      *sync.Mutex
	reset     chan time.Duration
	quit      chan bool
	waitGroup *sync.WaitGroup
}

type defaultJob struct {
	fct func()
}

func New() *Scheduler {
	return &Scheduler{
		scheduledJobs: make([]ScheduledJob, 0),
		waitGroup:     sync.WaitGroup{},
	}
}

func (scheduler *Scheduler) Schedule(scheduledJob ScheduledJob) {
	scheduledJob.jobControl.waitGroup = &scheduler.waitGroup
	scheduler.scheduledJobs = append(scheduler.scheduledJobs, scheduledJob)
}

func (scheduler *Scheduler) Run() {

	for _, job := range scheduler.scheduledJobs {
		go tickerRunner(job)
	}
}

func (scheduler *Scheduler) Wait() {
	scheduler.waitGroup.Wait()
}

func (job *ScheduledJob) Every(duration time.Duration) *ScheduledJob {
	return job.setDuration(duration)
}

func (job *ScheduledJob) Milliseconds(d int) *ScheduledJob {
	return job.plusDuration(time.Duration(d * int(time.Millisecond)))
}

func (job *ScheduledJob) Seconds(d int) *ScheduledJob {
	return job.plusDuration(time.Duration(d * int(time.Second)))
}

func (job *ScheduledJob) Minutes(d int) *ScheduledJob {
	return job.plusDuration(time.Duration(d * int(time.Minute)))
}

func (job *ScheduledJob) Hours(d int) *ScheduledJob {
	return job.plusDuration(time.Duration(d * int(time.Hour)))
}

func (job *ScheduledJob) Days(d int) *ScheduledJob {
	return job.Hours(d * 24)
}

func (job *ScheduledJob) setDuration(duration time.Duration) *ScheduledJob {
	job.Tickduration = duration
	job.jobControl.reset <- job.Tickduration
	return job
}

func (job *ScheduledJob) plusDuration(duration time.Duration) *ScheduledJob {
	return job.setDuration(job.Tickduration + duration)
}

func functionJobBuilder(fct func()) Job {
	return defaultJob{
		fct: fct,
	}
}

func tickerRunner(scheduledJob ScheduledJob) {

	defer scheduledJob.jobControl.waitGroup.Done()

	ticker := time.Ticker{}
	tick := ticker.C

	for {
		select {
		case <-tick:
			scheduledJob.jobControl.lock.Lock()
			scheduledJob.Job.Run()
			scheduledJob.jobControl.lock.Unlock()
		case newDuration := <-scheduledJob.jobControl.reset:
			ticker.Reset(newDuration)
		case <-scheduledJob.jobControl.quit:
			scheduledJob.jobControl.lock.Lock()
			ticker.Stop()
			return
		}
	}
}

func do(feedConfig *config.Feed) {

	feed, err := feed.Fetch(feedConfig)

	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Unable to fetch feed [%s] at URL [%s]", feedConfig.Name, feedConfig.Url))
	}

	err = feed.Save()

	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Unable to persist feed [%s] :\n%s", feedConfig.Name, err))
	}
}

func (job defaultJob) Run() {
	job.fct()
}
