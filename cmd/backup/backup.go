package backup

import (
	"backupgo/config"
	"backupgo/notice"
	"backupgo/oss"
	"backupgo/task"
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

func BackupCommand() *cli.Command {
	return &cli.Command{
		Name:  "backup",
		Usage: "Run a specific backup task manually",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args()
			if args.Len() == 0 {
				return fmt.Errorf("missing required argument: backup-id\nSee 'backupgo backup --help' for more information")
			}
			return runBackup(args.First())
		},
	}
}

func runBackup(backupID string) error {
	config.InitConfig()

	conf, ok := config.Config.FindBackupByID(backupID)
	if !ok {
		return fmt.Errorf("backup task not found: %s", backupID)
	}

	ossClient := oss.CreateOSSClient(config.Config.OSS)
	noticeManager := notice.NewManagerFromConfig(config.Config)

	holder := task.NewTaskHolder(conf, ossClient, noticeManager)

	fmt.Printf("Running backup task: %s\n", backupID)
	holder.BackupTask()
	fmt.Printf("Backup task completed: %s\n", backupID)

	return nil
}
