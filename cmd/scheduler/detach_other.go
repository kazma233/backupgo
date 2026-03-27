//go:build !unix

package scheduler

import (
	"errors"
	"os/exec"
)

func applyDetach(cmd *exec.Cmd) error {
	return errors.New("background mode is only supported on Unix-like systems")
}
