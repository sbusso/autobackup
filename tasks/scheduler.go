package tasks

import (
	"log"
	"math/rand"
	"time"

	"github.com/robfig/cron"
	"github.com/sbusso/autobackup/sources"
	"github.com/sbusso/autobackup/stores"
)

type Scheduler struct {
	cfg      *Config
	task     task
	cr       *cron.Cron
	schedule string
	quit     chan bool
}

func ScheduleBackup(c *Config, source sources.Source, store stores.Store) (*Scheduler, error) {
	s := NewScheduler(c, func(c *Config) error {
		return BackupTask(c, source, store)
	})
	return s, nil
}

func ScheduleRestore(c *Config, source sources.Source, store stores.Store) (*Scheduler, error) {
	s := NewScheduler(c, func(c *Config) error {
		return RestoreTask(c, source, store)
	})
	return s, nil
}

func NewScheduler(c *Config, task task) *Scheduler {
	return &Scheduler{
		cfg:      c,
		task:     task,
		cr:       cron.New(),
		schedule: c.Schedule,
	}
}

func (s *Scheduler) Stop() error {
	s.quit <- true
	return nil
}

func (s *Scheduler) Start() error {

	if s.schedule == "" || s.schedule == "none" {
		log.Println("Running task directly")
		return s.task(s.cfg)
	}

	log.Println("Starting scheduled backup task")
	s.quit = make(chan bool, 1)

	s.cr.AddFunc(s.schedule, func() {
		delay := s.cfg.RandomDelay
		if delay <= 0 {
			log.Println("Schedule random delay was set to a number <= 0, using 1 as default")
			delay = 1
		}

		seconds := rand.Intn(delay)

		// run immediately is no delay is configured
		if seconds == 0 {
			if err := s.task(s.cfg); err != nil {
				log.Printf("Failed to run scheduled task: %v\n", err)
			}
			return
		}

		log.Printf("Waiting for %d seconds before starting scheduled job", seconds)

		select {
		case <-s.quit:
			log.Println("Quiting")
			break
		case <-time.After(time.Duration(seconds) * time.Second):
			log.Println("Running scheduled task")

			if err := s.task(s.cfg); err != nil {
				log.Println(0, "Failed to run scheduled task: %v\n", err)
			}
			break
		}
	})

	go func() {
		s.cr.Start()

		<-s.quit
		close(s.quit)

		log.Println("Stopping scheduled task")
		s.cr.Stop()
	}()

	return nil
}
