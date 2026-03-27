//go:build !unix

package procutil

import (
	"errors"
	"os"
)

func IsRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	return process.Pid != 0, nil
}

func Terminate(pid int) error {
	return errors.New("terminate is only supported on Unix-like systems")
}
