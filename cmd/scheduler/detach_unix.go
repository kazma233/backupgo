//go:build unix

package scheduler

import (
	"os/exec"
	"syscall"
)

func applyDetach(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return nil
}
