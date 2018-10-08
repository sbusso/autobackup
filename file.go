package autobackup

import (
	"fmt"

	"github.com/sbusso/autobackup/sources"
	"github.com/sbusso/autobackup/stores"
	"github.com/sbusso/autobackup/tasks"
)

// File recipe to backup only one file
func File(dbName string) (*tasks.Scheduler, error) {

	var config = tasks.NewConfig()

	var opts = map[string]interface{}{
		"File": dbName,
	}

	var source = sources.NewTarballConfig(opts)

	var store = stores.NewS3Config()

	s, err := tasks.ScheduleBackup(config, source, store)
	if err != nil {
		return nil, fmt.Errorf("an error occured during backup: %v\n", err)
	}

	s.Start()

	return s, nil
}
