package sources

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"log"
)

// CmdConfig has the configuration needed to run an external command
type CmdConfig struct {
	Env        []string
	InputFile  io.Reader
	OutputFile io.Writer
	Credential *syscall.Credential
	ParsedArg  string
}

// CmdRun executes an external executable
func (app *CmdConfig) CmdRun(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Env = app.Env

	// only switch user when running as root
	if euid := os.Geteuid(); euid == 0 && app.Credential != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = app.Credential
	} else if euid != 0 {
		log.Printf("Not running as root, starting %s with UID %d\n", name, os.Geteuid())
	}

	if app.InputFile == nil && app.OutputFile == nil {
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	var readErr, writeErr error

	doneWrite := make(chan error)
	doneRead := make(chan error)

	if app.OutputFile != nil {
		outPipe, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("cannot create stdout pipe: %v", err)
		}

		reader := bufio.NewReader(outPipe)

		log.Println("Sending command stdout to file")

		go func() {
			_, err := io.Copy(app.OutputFile, reader)
			doneWrite <- err
		}()
	} else {
		cmd.Stdout = os.Stdout
		close(doneWrite)
	}

	if app.InputFile != nil {
		inPipe, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("cannot create stdin pipe: %v", err)
		}

		log.Println("Sending file to command stdin")

		go func() {
			_, err := io.Copy(inPipe, app.InputFile)
			inPipe.Close()
			doneRead <- err
		}()
	} else {
		close(doneRead)
	}

	log.Printf("Running %s %s\n", name, strings.Join(parseArg(arg, app.ParsedArg), " "))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot start process: %v", err)
	}

	writeErr = <-doneWrite
	readErr = <-doneRead

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for process: %v", err)
	}

	if readErr != nil {
		return fmt.Errorf("failed to read process stdin: %v", readErr)
	}

	if writeErr != nil {
		return fmt.Errorf("failed to write process stdout: %v", writeErr)
	}

	return nil
}

func generateFilename(dir, prefix string) string {
	now := time.Now().Format("20060102150405")
	return path.Join(dir, prefix+"-"+now)
}

func removeDirectoryContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("cannot open directory: %v", err)
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("cannot read files on directory: %v", err)
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("failed to remove %s: %v", name, err)
		}
	}

	return nil
}

func parseArg(args []string, arg string) []string {
	var updated []string

	if arg == "" {
		updated = args
		return updated
	}

	isShort := !strings.HasPrefix(arg, "--")
	for i, a := range args {
		if isShort {
			if strings.HasPrefix(a, arg) {
				updated = append(updated, arg+"********")
				updated = append(updated, args[i+1:]...)
				break
			}
		} else {
			if a == arg {
				updated = append(updated, arg, "*****")
				updated = append(updated, args[i+2:]...)
				break
			}
		}

		updated = append(updated, a)
	}

	return updated
}
