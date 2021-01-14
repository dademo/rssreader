package scheduler

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dademo/rssreader/modules/config"
	"github.com/dademo/rssreader/modules/feed"

	log "github.com/sirupsen/logrus"
)

type Scheduler struct {
	scheduledJobs []ScheduledJob
	wait          sync.WaitGroup
}

type Job interface {
	Run()
}

type ScheduledJob struct {
	Job                      Job
	TickDurationMilliseconds time.Duration
	jobControl               *jobControl
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
		wait:          sync.WaitGroup{},
	}
}

func (scheduler *Scheduler) Schedule(scheduledJob ScheduledJob) {
	scheduler.scheduledJobs = append(scheduler.scheduledJobs, scheduledJob)
	// scheduler.AddJob(
	// 	makeSpec(feedConfig),
	// 	FeedFetchJob{feedConfig: feedConfig},
	// )
}

func (scheduler *Scheduler) Run() {

	for _, job := range scheduler.scheduledJobs {
		go tickerRunner(job)
	}
}

func (scheduler *Scheduler) Wait() {
	scheduler.wait.Wait()
}

func (job *ScheduledJob) Every(duration time.Duration) *ScheduledJob {
	return job.setDuration(duration * 1000 * 1000)
}

func (job *ScheduledJob) Milliseconds(duration int) *ScheduledJob {
	return job.plusDuration(time.Duration(duration * int(math.Pow10(6))))
}

func (job *ScheduledJob) Seconds(duration int) *ScheduledJob {
	return job.plusDuration(time.Duration(duration * int(math.Pow10(9))))
}

func (job *ScheduledJob) Minutes(duration int) *ScheduledJob {
	return job.Seconds(duration * 60)
}

func (job *ScheduledJob) Hours(duration int) *ScheduledJob {
	return job.Minutes(duration * 60)
}

func (job *ScheduledJob) Days(duration int) *ScheduledJob {
	return job.Hours(duration * 24)
}

func (job *ScheduledJob) setDuration(duration time.Duration) *ScheduledJob {
	job.TickDurationMilliseconds = duration
	job.jobControl.reset <- job.TickDurationMilliseconds
	return job
}

func (job *ScheduledJob) plusDuration(duration time.Duration) *ScheduledJob {
	return job.setDuration(job.TickDurationMilliseconds + duration)
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
			ticker.Stop()
			scheduledJob.jobControl.lock.Lock()
			return
		}
	}
}

func do(feedConfig config.Feed) {

	feed, err := feed.Fetch(feedConfig)

	if err != nil {
		log.Error(fmt.Sprintf("Unable to fetch feed [%s] at URL [%s]", feedConfig.Name, feedConfig.Url))
	}

	err = feed.Save()

	if err != nil {
		log.Error(fmt.Sprintf("Unable to persist feed [%s] :\n%s", feedConfig.Name, err))
	}
}

func (job defaultJob) Run() {
	job.fct()
}
