package tasks

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
	"github.com/sbusso/autobackup/sources"
	"github.com/sbusso/autobackup/stores"
)

// Config store global information required for backup tasks
type Config struct {
	// App           *App
	// Command       Command
	// shellComplete bool
	// flagSet       *flag.FlagSet
	// setFlags      map[string]bool
	Schedule    string `env:"SCHEDULE" envDefault:"@daily"`
	MaxBackups  int    `env:"MAX_BACKUPS" envDefault:"7"`
	RestoreFile string `env:"RESTORE_FILE"`
	RandomDelay int    `env:"RANDOM_DELAY" envDefault:"1"`
}

func NewConfig() *Config {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}

type task func(c *Config) error

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func ScheduleBackup(c *Config, source sources.Source, store stores.Store) error {
	return runScheduler(c, func(c *Config) error {
		return BackupTask(c, source, store)
	})
}

func ScheduleRestore(c *Config, source sources.Source, store stores.Store) error {
	return runScheduler(c, func(c *Config) error {
		return RestoreTask(c, source, store)
	})
}

func BackupTask(c *Config, source sources.Source, store stores.Store) error {
	filepath, err := source.Backup()
	if err != nil {
		return fmt.Errorf("source backup failed: %v", err)
	}

	log.Printf("Backup saved to %s\n", filepath)

	filename := path.Base(filepath)

	if err = store.Store(filepath, filename); err != nil {
		return fmt.Errorf("couldn't upload file to store: %v", err)
	}

	err = store.RemoveOlderBackups(c.MaxBackups)
	if err != nil {
		return fmt.Errorf("couldn't remove old backups from store: %v", err)
	}

	return nil
}

func RestoreTask(c *Config, source sources.Source, store stores.Store) error {
	var err error
	var filename string

	if key := c.RestoreFile; key != "" {
		// restore directly from this file
		filename = key
	} else {
		// find the latest file in the store
		filename, err = store.FindLatestBackup()
		if err != nil {
			return fmt.Errorf("cannot find the latest backup: %v", err)
		}
	}

	filepath, err := store.Retrieve(filename)
	if err != nil {
		return fmt.Errorf("cannot download file %s: %v", filename, err)
	}

	defer store.Close()

	if err = source.Restore(filepath); err != nil {
		return fmt.Errorf("source restore failed: %v", err)
	}

	return nil
}

func runScheduler(c *Config, task task) error {
	cr := cron.New()
	schedule := c.Schedule

	if schedule == "" || schedule == "none" {
		log.Println("Running task directly")
		return task(c)
	}

	log.Println("Starting scheduled backup task")
	timeoutchan := make(chan bool, 1)

	cr.AddFunc(schedule, func() {
		delay := c.RandomDelay
		if delay <= 0 {
			log.Println("Schedule random delay was set to a number <= 0, using 1 as default")
			delay = 1
		}

		seconds := rand.Intn(delay)

		// run immediately is no delay is configured
		if seconds == 0 {
			if err := task(c); err != nil {
				log.Printf("Failed to run scheduled task: %v\n", err)
			}
			return
		}

		log.Printf("Waiting for %d seconds before starting scheduled job", seconds)

		select {
		case <-timeoutchan:
			log.Println("Random timeout cancelled")
			break
		case <-time.After(time.Duration(seconds) * time.Second):
			log.Println("Running scheduled task")

			if err := task(c); err != nil {
				log.Println(0, "Failed to run scheduled task: %v\n", err)
			}
			break
		}
	})
	cr.Start()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	timeoutchan <- true
	close(timeoutchan)

	log.Println("Stopping scheduled task")
	cr.Stop()

	return nil
}
