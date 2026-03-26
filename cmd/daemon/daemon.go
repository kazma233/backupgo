package daemon

import (
	"backupgo/config"
	"backupgo/notice"
	"backupgo/oss"
	"backupgo/pkg/consts"
	"backupgo/task"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"
	"github.com/urfave/cli/v3"
)

var cronScheduler *cron.Cron

func DaemonCommand() *cli.Command {
	return &cli.Command{
		Name:  "daemon",
		Usage: "Start backupgo as a daemon (background process)",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runDaemon()
		},
	}
}

func runDaemon() error {
	config.InitConfig()

	ossClient := oss.CreateOSSClient(config.Config.OSS)
	noticeManager := notice.NewManagerFromConfig(config.Config)

	cronScheduler = cron.New(cron.WithParser(cron.NewParser(
		cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.DowOptional|cron.Descriptor,
	)), cron.WithChain())

	for _, conf := range config.Config.BackupConf {
		holder := task.NewTaskHolder(conf, ossClient, noticeManager)

		backupTaskCron := conf.BackupTask
		if backupTaskCron == "" {
			backupTaskCron = "0 25 0 * * ?"
		}
		_, err := cronScheduler.AddFunc(backupTaskCron, func() {
			holder.BackupTask()
		})
		if err != nil {
			return err
		}

		log.Printf("task %s added to scheduler", conf.GetID())
	}

	if err := writePID(); err != nil {
		log.Printf("failed to write PID file: %v", err)
	}

	log.Println("backupgo daemon started")

	cronScheduler.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan

	log.Println("shutting down...")
	cronScheduler.Stop()
	removePID()

	return nil
}

func writePID() error {
	pid := os.Getpid()
	return os.WriteFile(consts.PIDFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func removePID() {
	os.Remove(consts.PIDFile)
}
