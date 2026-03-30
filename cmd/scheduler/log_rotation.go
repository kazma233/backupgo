package scheduler

import (
	"fmt"
	"os"
)

func rotateLogFile(logFilePath string, backupFilePath string) error {
	if _, err := os.Stat(logFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat log file failed: %w", err)
	}

	if err := os.Remove(backupFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old log backup failed: %w", err)
	}

	if err := os.Rename(logFilePath, backupFilePath); err != nil {
		return fmt.Errorf("backup log file failed: %w", err)
	}

	return nil
}
