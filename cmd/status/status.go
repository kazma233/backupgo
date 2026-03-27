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
	showPID()
	fmt.Println()
	listTasks()
	return nil
}

func showPID() {
	pid, err := readPID()
	if err != nil {
		fmt.Println("Scheduler status: not running (no PID file)")
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

func readPID() (int, error) {
	pidFile, err := consts.PIDFilePath()
	if err != nil {
		return 0, err
	}

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
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
