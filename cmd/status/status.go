package status

import (
	"backupgo/config"
	"backupgo/pkg/consts"
	"backupgo/pkg/procutil"
	"backupgo/state"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Show scheduler status and list all backup tasks",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runStatus()
		},
	}
}

func runStatus() error {
	pidFile, err := consts.PIDFilePath()
	if err != nil {
		return err
	}

	logFile, err := consts.LogFilePath()
	if err != nil {
		return err
	}

	stateFile, err := consts.StateFilePath()
	if err != nil {
		return err
	}

	showPID(pidFile)
	fmt.Printf("PID file: %s\n", pidFile)
	fmt.Printf("Log file: %s\n", logFile)
	fmt.Printf("State file: %s\n", stateFile)
	fmt.Println()
	listTasks()
	return nil
}

func showPID(pidFile string) {
	pid, err := readPID(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Scheduler status: not running (no PID file)")
			return
		}

		fmt.Printf("Scheduler status: unknown (read PID file failed: %v)\n", err)
		return
	}

	running, err := procutil.IsRunning(pid)
	if err != nil {
		fmt.Printf("Scheduler status: PID %d (process check failed: %v)\n", pid, err)
		return
	}

	if !running {
		fmt.Printf("Scheduler status: not running (stale PID file: %d)\n", pid)
		return
	}

	fmt.Printf("Scheduler status: running (PID %d)\n", pid)
}

func readPID(pidFile string) (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid file %s failed: %w", pidFile, err)
	}

	return pid, nil
}

func listTasks() {
	config.InitConfig()

	fmt.Println("Backup tasks:")
	fmt.Println("-------------------------------------------------------------------")

	format := "%-20s %-12s %-20s %s\n"
	fmt.Printf(format, "ID", "TYPE", "CRON", "LAST RUN")

	for _, conf := range config.Config.BackupConf {
		cronExpr := conf.BackupTask
		if cronExpr == "" {
			cronExpr = "0 25 0 * * ? (default)"
		}

		taskState := state.GetState().GetTaskState(conf.GetID())
		lastRun := "never"
		if taskState != nil {
			lastRun = taskState.LastRun.Format(time.RFC3339)
		}

		fmt.Printf(format, conf.GetID(), conf.GetType(), cronExpr, lastRun)
	}
}
