package tasks

import (
	"fmt"
	"log"
	"path"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
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
