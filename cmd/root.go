package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"backupgo/cmd/backup"
	"backupgo/cmd/scheduler"
	"backupgo/cmd/status"
)

func Run(args []string) error {
	rootCmd := &cli.Command{
		Name:  "backupgo",
		Usage: "Backup management tool",
		Commands: []*cli.Command{
			scheduler.StartCommand(),
			scheduler.StopCommand(),
			status.StatusCommand(),
			backup.BackupCommand(),
		},
	}

	return rootCmd.Run(context.Background(), args)
}
