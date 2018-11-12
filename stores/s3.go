package stores

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/caarlos0/env"
)

// S3Config has the config options for the S3 service
type S3Config struct {
	Endpoint        string `env:"S3_ENDPOINT"`
	Region          string `env:"S3_REGION"`
	Bucket          string `env:"S3_BUCKET"`
	Prefix          string `env:"S3_PREFIX"`
	ForcePathStyle  bool   `env:"S3_FORCE_PATH_STYLE" envDefault:"false"`
	KeepAfterUpload bool   `env:"KEEP_AFTER_UPLOAD" envDefault:"false"`
	SaveDir         string `env:"SAVEDIR" envDefault:"/tmp/"`
	retrievedFile   string `env:"RETRIEVED_FILE"`
}

func NewS3Config() (*S3Config, error) {
	cfg := &S3Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *S3Config) newSession() *session.Session {
	config := &aws.Config{
		Endpoint:         aws.String(s.Endpoint),
		Region:           aws.String(s.Region),
		S3ForcePathStyle: aws.Bool(s.ForcePathStyle),
	}

	return session.Must(session.NewSession(config))
}

// Store saves a file to a remote S3 service
func (s *S3Config) Store(filepath string, filename string) error {
	uploader := s3manager.NewUploader(s.newSession())

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", filepath, err)
	}

	defer f.Close()

	if !s.KeepAfterUpload {
		defer func() {
			log.Printf("Removing source file %s\n", filepath)
			if err = os.Remove(filepath); err != nil {
				log.Printf("Cannot remove file %s, %v\n", filepath, err)
			}
		}()
	}

	key := path.Clean(path.Join(s.Prefix, filename))

	// Upload the file to S3.
	res, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	log.Printf("File uploaded to %s\n", res.Location)

	return nil
}

func (s *S3Config) getFileListing(svc *s3.S3) ([]string, error) {
	var files []string

	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket),
		// make sure that the prefix ends with "/"
		Prefix: aws.String(path.Clean(s.Prefix) + "/"),
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {

		for _, obj := range p.Contents {
			if !strings.HasSuffix(*obj.Key, "/") {
				files = append(files, aws.StringValue(obj.Key))
			}
		}
		return true
	})

	return files, err
}

// RemoveOlderBackups keeps the most recent backups of the S3 service and deletes the old ones
func (s *S3Config) RemoveOlderBackups(keep int) error {
	svc := s3.New(s.newSession())

	files, err := s.getFileListing(svc)
	if err != nil {
		return fmt.Errorf("couldn't list S3 objects, %v", err)
	}

	sort.Strings(files)
	count := len(files) - keep

	if count > 0 {
		var items s3.Delete
		var objs = make([]*s3.ObjectIdentifier, count)

		for i, file := range files[:count] {
			objs[i] = &s3.ObjectIdentifier{Key: aws.String(file)}
			log.Printf("Marked to delete: s3://%s/%s\n", s.Bucket, file)
		}

		items.SetObjects(objs)

		out, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(s.Bucket),
			Delete: &items})

		if err != nil {
			return fmt.Errorf("couldn't delete the S3 objects, %v", err)
		}

		log.Printf("Deleted %d objects from S3\n", len(out.Deleted))
	}

	return nil
}

// FindLatestBackup returns the most recent backup of the S3 store
func (s *S3Config) FindLatestBackup() (string, error) {
	svc := s3.New(s.newSession())

	files, err := s.getFileListing(svc)
	if err != nil {
		return "", fmt.Errorf("couldn't list S3 objects, %v", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("cannot find a recent backup on s3://%s/%s",
			s.Bucket, s.Prefix)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	return files[0], nil
}

// Retrieve downloads a S3 object to the local filesystem
func (s *S3Config) Retrieve(s3path string) (string, error) {
	// Create an uploader with the session and default options
	downloader := s3manager.NewDownloader(s.newSession())

	filepath := path.Join(s.SaveDir, path.Base(s3path))
	f, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}

	defer f.Close()

	// download the file from S3.
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s3path),
	})

	if err != nil {
		return "", fmt.Errorf("failed to download S3 object, %v", err)
	}

	log.Printf("File downloaded to %s\n", filepath)
	s.retrievedFile = filepath

	return filepath, nil
}

// Close deinitializes the store (remove downloaded file)
func (s *S3Config) Close() {
	if s.retrievedFile != "" {
		if err := os.Remove(s.retrievedFile); err != nil {
			log.Printf("Cannot remove file %s\n", s.retrievedFile)
		}

		s.retrievedFile = ""
	}
}
