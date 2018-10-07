package autobackup

import (
	"log"

	"github.com/sbusso/autobackup/sources"
	"github.com/sbusso/autobackup/stores"
	"github.com/sbusso/autobackup/tasks"
)

// File recipe to backup only one file
func File(dbName string) {

	var config = tasks.NewConfig()

	var opts = map[string]interface{}{
		"File": dbName,
	}

	var source = sources.NewTarballConfig(opts)

	var store = stores.NewS3Config()

	if err := tasks.ScheduleBackup(config, source, store); err != nil {
		log.Printf("an error occured during backup: %v\n", err)
		return
	}
}
