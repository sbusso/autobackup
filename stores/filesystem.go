package stores

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"log"
)

// FilesystemConfig has the config options for the FilesystemConfig service
type FilesystemConfig struct {
	SaveDir string
}

// Store moves/copies a file to another directory
func (f *FilesystemConfig) Store(src string, filename string) error {
	dest := path.Clean(path.Join(f.SaveDir, filename))

	if src == dest {
		log.Println("Using the same path as source and destination, do nothing")
		return nil
	}

	err := os.Rename(dest, src)
	if err != nil {
		log.Printf("Cannot rename %s to %s, trying to copy instead\n", src, dest)
	} else {
		return nil
	}

	var removeSourceFile = false

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("cannot open source file %s, %v", src, err)
	}

	defer func() {
		srcFile.Close()

		// if there aren't any errors on the file copy then delete the source file
		if removeSourceFile {
			if err = os.Remove(src); err != nil {
				log.Printf("Cannot remove source file %s\n", src)
			}
		}
	}()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("cannot create destination file %s, %v", dest, err)
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	if err != nil {
		return fmt.Errorf("error while copying file, %v", err)
	}

	if err = destFile.Sync(); err != nil {
		return fmt.Errorf("cannot flush file contents, %v", err)
	}

	removeSourceFile = true

	return nil
}

// RemoveOlderBackups keeps the most recent backups of a directory and deletes the old ones
func (f *FilesystemConfig) RemoveOlderBackups(keep int) error {
	files, err := ioutil.ReadDir(f.SaveDir)
	if err != nil {
		return fmt.Errorf("cannot list contents of directory %s, %v", f.SaveDir, err)
	}

	count := len(files) - keep
	deleted := 0

	if count > 0 {
		for _, file := range files[:count] {
			fullpath := path.Clean(path.Join(f.SaveDir, file.Name()))
			err = os.Remove(fullpath)
			if err != nil {
				log.Println(0, "Failed to remove file %s", fullpath)
			} else {
				deleted++
			}
		}

		log.Printf("Deleted %d objects from %s\n", deleted, f.SaveDir)
	}

	return nil
}

// FindLatestBackup returns the most recent backup of the specified directory
func (f *FilesystemConfig) FindLatestBackup() (string, error) {
	files, err := ioutil.ReadDir(f.SaveDir)
	if err != nil {
		return "", fmt.Errorf("cannot list contents of directory %s, %v", f.SaveDir, err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("cannot find a recent backup on %s", f.SaveDir)
	}

	return files[len(files)-1].Name(), nil
}

// Retrieve returns the path of the requested file
func (f *FilesystemConfig) Retrieve(filename string) (string, error) {
	return path.Clean(path.Join(f.SaveDir, filename)), nil
}

// Close deinitializes the store (no dothing)
func (f *FilesystemConfig) Close() {
}
