package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_zipPath(t *testing.T) {
	sourceDir := t.TempDir()
	target := filepath.Join(t.TempDir(), "test.zip")

	filePath := filepath.Join(sourceDir, "demo.txt")
	if err := os.WriteFile(filePath, []byte("backupgo test"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	callbackCalled := false
	doneCalled := false
	path, err := ZipPath(sourceDir, target, func(filePath string, processed, total int64, percentage float64) {
		callbackCalled = callbackCalled || total >= 0 || processed >= 0 || percentage >= 0 || filePath == ""
	}, func(total int64) {
		doneCalled = total > 0
	})
	if err != nil {
		t.Fatalf("zip path: %v", err)
	}

	if path != target {
		t.Fatalf("expected target %q, got %q", target, path)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("stat zip file: %v", err)
	}
	if !doneCalled {
		t.Fatalf("expected done callback to be called")
	}
	if callbackCalled {
		t.Log("progress callback executed during test")
	}
}
