# AutoBackup

>_Easy backup system directly from your binary_

The primary advantage of Golang is to produce no dependency binaries you can deploy independently to different platforms. With Kubernetes or other container architectures, backup is managed by some cloud orchestrator, but in the case of non-cloud architecture (and Golang is very good for that purpose too) you need to setup a backup solution, and it should not be optional if you store data. To keep your application independent and easy to deploy / maintain, autobackup provides methods to your application to run and schedule backup and restore tasks. It is as simple as adding this line of code `autobackup.File("products.db")` and _voila_ you have scheduled a daily backup of your bolt database to an S3 compatible storage. No more excuses to not backup your data !

This module can be used to make regular backups of various applications and restores from S3. It will perform daily backups unless configured to do otherwise.

## Usage

Using an existing recipe:

``` go
import "github.com/sbusso/autobackup"

autobackup.File("products.db")

```

More advance usage:

``` go
import (
  "log"

  "github.com/sbusso/autobackup/sources"
  "github.com/sbusso/autobackup/stores"
  "github.com/sbusso/autobackup/tasks"
)

var config = tasks.NewConfig()

var opts = map[string]interface{}{
  "File": dbName,
}

var service = sources.NewTarballConfig(opts)

var store = stores.NewS3Config()

if err := tasks.BackupTask(config, service, store); err != nil {
  log.Printf("an error occurred during backup: %v\n", err)
  return
}
```

### Supported sources

* PostgreSQL
* MySQL
* Tarball
* Consul

### Supported stores

* S3
* Filesystem (local)

The schedule function can also be used on restore if you need to test your backups regularly.

## Sources Configuration

Those configuration variables can be setup by environment, .env files, or when building an advanced task for backup.

### Backup and restore

* `SAVE_DIR`: directory to store the temporal backup after creating/retrieving it.`
* `SCHEDULE_RANDOM_DELAY`: maximum number of seconds (value chosen at random) to wait before starting a task. There is no random delay by default.
* `SCHEDULE`: specifies when to start a task. Defaults to `@daily` on backup, `none` on restore. Accepts cron format, like `0 0 * * *`. Set to `none` to disable and perform only one task.

### Backup only

* `MAX_BACKUPS`: maximum number of backups to keep on the store.

### Restore only

* `RESTORE_FILE`: Restore directly from this filename instead of searching for the most recent one. Only used with the `restore` command.

### Database

* `DATABASE_HOST`: database host.
* `DATABASE_PORT`: database port.
* `DATABASE_NAME`: database name.
* `DATABASE_USER`:  database user.
* `DATABASE_PASSWORD`:  database password.
* `DATABASE_PASSWORD_FILE`:  database password file, has precendnce over `DATABASE_PASSWORD`
* `DATABASE_OPTIONS`:  custom options to pass to the backup/restore application.
* `DATABASE_COMPRESS`: compress the sql file with gzip.
* `DATABASE_IGNORE_EXIT_CODE`: ignore is the restore operation returns a non-zero exit code.

### Postgres

* `POSTGRES_CUSTOM_FORMAT`: use custom dump format instead of plain text backups.

### Tarball

* `TARBALL_FILE`: file to backup/restore relative to specified PATH.
* `TARBALL_PATH`: directory to backup/restore default is current directory.
* `TARBALL_NAME_PREFIX`: name prefix of the created tarball. If unset it will use the backup directory name.
* `TARBALL_COMPRESS`: compress the tarball with gzip default is `true`.

## S3 Configuration

* `S3_ENDPOINT`: url of the S3 compatible endpoint, for example `https://nyc3.digitaloceanspaces.com`.
* `S3_REGION`: region where the bucket is located, for example `us-east-1`.
* `S3_BUCKET`: name of the bucket, for example `backups`.
* `S3_PREFIX`: for example `private/files`.
* `S3_FORCE_PATH_STYLE`: set to `1` if you are using minio.
* `S3_KEEP_FILE`: keep file on the local filesystem after uploading it to S3.

The credentials are passed using the standard variables:

* `AWS_ACCESS_KEY_ID`: AWS access key. `AWS_ACCESS_KEY` can also be used.
* `AWS_SECRET_ACCESS_KEY`: AWS secret key. `AWS_SECRET_KEY` can also be used.
* `AWS_SESSION_TOKEN`: AWS session token. Optional, will be used if present.

## TODO

* [ ] tests, more tests, even more tests
* [ ] add checksum to backups and check them when restoring

## License and Copyright

The core code for BackupTask, RestoreTask, Sources (Services) and Stores is extracted from [codestation/go-s3-backup](https://github.com/codestation/go-s3-backup) copyright by Codestation and licensed under the Apache License 2.0. Due to original design and purpose of the application, it couldn't be forked or imported, thus extracted code has been restructured and refactored to remove the command line interface and its dependencies, to simplify configuration management, to adopt a different naming convention and to change some backup behavior. This has permitted to get backup and restore behaviors with an embedded interface instead of command line. Gogs service has not been imported.

Other part of the code is licensed under the same Apache License 2.0 and copyright by Stephane Busso.
