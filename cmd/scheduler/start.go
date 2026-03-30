package scheduler

import (
	"backupgo/config"
	"backupgo/notice"
	"backupgo/oss"
	"backupgo/pkg/consts"
	"backupgo/pkg/procutil"
	"backupgo/task"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v3"
)

var cronScheduler *cron.Cron

func StartCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start backupgo scheduler",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "detach",
				Aliases:     []string{"d"},
				Usage:       "Run in background",
				HideDefault: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runStart(cmd.Bool("detach"))
		},
	}
}

func runStart(detach bool) error {
	if detach {
		return runDetached()
	}

	return runForeground()
}

func runForeground() error {
	if err := ensureNotRunning(); err != nil {
		return err
	}

	config.InitConfig()

	ossClient := oss.CreateOSSClient(config.Config.OSS)
	noticeManager := notice.NewManagerFromConfig(config.Config)

	cronScheduler = cron.New(cron.WithParser(cron.NewParser(
		cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.DowOptional|cron.Descriptor,
	)), cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))

	for _, conf := range config.Config.BackupConf {
		backupTaskCron := conf.BackupTask
		if backupTaskCron == "" {
			backupTaskCron = "0 25 0 * * ?"
		}
		_, err := cronScheduler.AddFunc(backupTaskCron, func() {
			holder := task.NewTaskHolder(conf, ossClient, noticeManager)
			holder.BackupTask()
		})
		if err != nil {
			return err
		}

		log.Printf("task %s added to scheduler", conf.GetID())
	}

	cronScheduler.Start()
	if err := writePID(); err != nil {
		return err
	}
	defer removePID()

	log.Println("backupgo scheduler started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigChan)
	<-sigChan

	log.Println("shutting down...")
	cronScheduler.Stop()

	return nil
}

func runDetached() error {
	if err := ensureNotRunning(); err != nil {
		return err
	}

	if _, err := consts.EnsureStateDir(); err != nil {
		return err
	}

	logFilePath, err := consts.LogFilePath()
	if err != nil {
		return err
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}
	defer logFile.Close()

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("open %s failed: %w", os.DevNull, err)
	}
	defer devNull.Close()

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable failed: %w", err)
	}

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory failed: %w", err)
	}

	cmd := exec.Command(executable, "start")
	cmd.Dir = workdir
	cmd.Env = os.Environ()
	cmd.Stdin = devNull
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := applyDetach(cmd); err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start background process failed: %w", err)
	}

	backgroundPID := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("release background process handle failed: %w", err)
	}

	if err := waitForPIDFile(backgroundPID, 5*time.Second); err != nil {
		return fmt.Errorf("%w (log: %s)", err, logFilePath)
	}

	fmt.Printf("backupgo started in background (PID %d)\n", backgroundPID)
	fmt.Printf("log file: %s\n", logFilePath)
	return nil
}

func writePID() error {
	if _, err := consts.EnsureStateDir(); err != nil {
		return err
	}

	pidFile, err := consts.PIDFilePath()
	if err != nil {
		return err
	}

	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func removePID() error {
	pidFile, err := consts.PIDFilePath()
	if err != nil {
		return err
	}

	return os.Remove(pidFile)
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

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse pid file %s failed: %w", filepath.Clean(pidFile), err)
	}

	return pid, nil
}

func ensureNotRunning() error {
	pid, err := readPID()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	running, err := procutil.IsRunning(pid)
	if err != nil {
		return fmt.Errorf("check process %d failed: %w", pid, err)
	}
	if running {
		return fmt.Errorf("backupgo is already running (PID %d)", pid)
	}

	if err := removePID(); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func waitForPIDFile(expectedPID int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		pid, err := readPID()
		if err == nil {
			if pid == expectedPID {
				running, runErr := procutil.IsRunning(pid)
				if runErr != nil {
					return fmt.Errorf("check process %d failed: %w", pid, runErr)
				}
				if running {
					return nil
				}
			}
		}

		running, err := procutil.IsRunning(expectedPID)
		if err != nil {
			return fmt.Errorf("check process %d failed: %w", expectedPID, err)
		}
		if !running {
			return fmt.Errorf("background process %d exited before startup completed", expectedPID)
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("background process %d did not create PID file within %s", expectedPID, timeout)
}
