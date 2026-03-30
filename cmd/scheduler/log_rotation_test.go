package scheduler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotateLogFileNoCurrentLog(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "backupgo.log")
	backupFilePath := filepath.Join(tempDir, "backupgo.log.bak")

	if err := rotateLogFile(logFilePath, backupFilePath); err != nil {
		t.Fatalf("rotateLogFile() error = %v", err)
	}

	if _, err := os.Stat(backupFilePath); !os.IsNotExist(err) {
		t.Fatalf("expected no backup file, got err = %v", err)
	}
}

func TestRotateLogFileReplacesOldBackup(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "backupgo.log")
	backupFilePath := filepath.Join(tempDir, "backupgo.log.bak")

	if err := os.WriteFile(logFilePath, []byte("current log"), 0644); err != nil {
		t.Fatalf("write current log failed: %v", err)
	}
	if err := os.WriteFile(backupFilePath, []byte("old backup"), 0644); err != nil {
		t.Fatalf("write old backup failed: %v", err)
	}

	if err := rotateLogFile(logFilePath, backupFilePath); err != nil {
		t.Fatalf("rotateLogFile() error = %v", err)
	}

	if _, err := os.Stat(logFilePath); !os.IsNotExist(err) {
		t.Fatalf("expected current log to be moved away, got err = %v", err)
	}

	backupContent, err := os.ReadFile(backupFilePath)
	if err != nil {
		t.Fatalf("read backup log failed: %v", err)
	}
	if got := string(backupContent); got != "current log" {
		t.Fatalf("backup content = %q, want %q", got, "current log")
	}
}
