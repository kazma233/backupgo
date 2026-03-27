package scheduler

import (
	"backupgo/pkg/procutil"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

func StopCommand() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop backupgo background process",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runStop()
		},
	}
}

func runStop() error {
	pid, err := readPID()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("backupgo is not running")
			return nil
		}
		return err
	}

	running, err := procutil.IsRunning(pid)
	if err != nil {
		return fmt.Errorf("check process %d failed: %w", pid, err)
	}
	if !running {
		if err := removePID(); err != nil && !os.IsNotExist(err) {
			return err
		}
		fmt.Printf("backupgo is not running, removed stale PID file for PID %d\n", pid)
		return nil
	}

	if err := procutil.Terminate(pid); err != nil {
		return fmt.Errorf("stop process %d failed: %w", pid, err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		running, err = procutil.IsRunning(pid)
		if err != nil {
			return fmt.Errorf("check process %d failed: %w", pid, err)
		}
		if !running {
			if err := removePID(); err != nil && !os.IsNotExist(err) {
				return err
			}
			fmt.Printf("backupgo stopped (PID %d)\n", pid)
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("process %d did not stop within 10 seconds", pid)
}
