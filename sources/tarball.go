package sources

import (
	"fmt"
	"path"

	"github.com/caarlos0/env"
	"github.com/mholt/archiver"
	"github.com/mitchellh/mapstructure"
)

// TarballConfig has the config options for the TarballConfig service
type TarballConfig struct {
	Name     string
	File     string `env:"TAR_FILE"`
	Path     string `env:"TAR_PATH" envDefault:"./"`
	Compress bool   `env:"TAR_COMPRESS" envDefault:"true"`
	SaveDir  string `env:"SAVEDIR" envDefault:"/tmp/"`
}

func NewTarballConfig(opts map[string]interface{}) *TarballConfig {
	cfg := &TarballConfig{}
	err := env.Parse(cfg)
	mapstructure.Decode(opts, cfg)

	if err != nil {
		fmt.Printf("%+v\n", err)
	}
	return cfg
}

func (f *TarballConfig) target() string {
	var target = f.Path

	if f.File != "" {
		target = target + f.File
	}

	return target
}

// Backup creates a tarball of the specified directory
func (f *TarballConfig) Backup() (string, error) {
	var name string
	var target = f.target()

	if f.Name != "" {
		name = f.Name + "-backup"
	} else {
		name = path.Base(target) + "-backup"
	}

	filepath := generateFilename(f.SaveDir, name) + ".tar"

	var err error

	if f.Compress {
		filepath += ".gz"
		err = archiver.TarGz.Make(filepath, []string{target})
	} else {
		err = archiver.Tar.Make(filepath, []string{target})
	}

	if err != nil {
		return "", fmt.Errorf("cannot create tarball on %s, %v", filepath, err)
	}

	return filepath, nil
}

// Restore extracts a tarball to the specified directory
func (f *TarballConfig) Restore(filepath string) error {
	err := removeDirectoryContents(f.target())
	if err != nil {
		return fmt.Errorf("failed to empty directory contents before restoring: %v", err)
	}

	archive := archiver.MatchingFormat(filepath)
	if archive == nil {
		return fmt.Errorf("unsupported file extension: %s", path.Base(filepath))
	}

	// use the parent directory to unpack as the current directory is already in the tarball
	err = archive.Open(filepath, path.Dir(f.Path))
	if err != nil {
		return fmt.Errorf("cannot unpack backup: %v", err)
	}

	return nil
}
