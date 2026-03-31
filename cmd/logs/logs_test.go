package logs

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyLogTail(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "backupgo.log")
	content := "line 1\nline 2\nline 3\n"
	want := "line 2\nline 3\n"
	if err := os.WriteFile(logFilePath, []byte(content), 0644); err != nil {
		t.Fatalf("write log file failed: %v", err)
	}

	var output bytes.Buffer
	if err := copyLogTail(&output, logFilePath, 2); err != nil {
		t.Fatalf("copyLogTail() error = %v", err)
	}

	if got := output.String(); got != want {
		t.Fatalf("copyLogTail() output = %q, want %q", got, want)
	}
}

func TestCopyLogTailMissing(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "missing.log")

	var output bytes.Buffer
	err := copyLogTail(&output, logFilePath, 100)
	if err == nil {
		t.Fatal("copyLogTail() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "open log file") {
		t.Fatalf("copyLogTail() error = %v, want open log file error", err)
	}
}

func TestCopyLogTailCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		lineCount int
		want      []string
	}{
		{
			name:      "last lines",
			input:     "line 1\nline 2\nline 3\n",
			lineCount: 2,
			want:      []string{"line 2\n", "line 3\n"},
		},
		{
			name:      "more than file length",
			input:     "line 1\nline 2\n",
			lineCount: 5,
			want:      []string{"line 1\n", "line 2\n"},
		},
		{
			name:      "without trailing newline",
			input:     "line 1\nline 2",
			lineCount: 1,
			want:      []string{"line 2"},
		},
		{
			name:      "zero lines",
			input:     "line 1\nline 2\n",
			lineCount: 0,
			want:      nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			logFilePath := filepath.Join(tempDir, "backupgo.log")
			if err := os.WriteFile(logFilePath, []byte(tt.input), 0644); err != nil {
				t.Fatalf("write log file failed: %v", err)
			}

			var output bytes.Buffer
			if err := copyLogTail(&output, logFilePath, tt.lineCount); err != nil {
				t.Fatalf("copyLogTail() error = %v", err)
			}

			got := output.String()
			want := strings.Join(tt.want, "")
			if got != want {
				t.Fatalf("copyLogTail() = %q, want %q", got, want)
			}
		})
	}
}

func TestCopyLogTailAcrossBlocks(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "backupgo.log")

	var content strings.Builder
	var want strings.Builder
	for i := 0; i < 300; i++ {
		line := fmt.Sprintf("line-%03d %s\n", i, strings.Repeat("x", 20))
		content.WriteString(line)
		if i >= 297 {
			want.WriteString(line)
		}
	}

	if err := os.WriteFile(logFilePath, []byte(content.String()), 0644); err != nil {
		t.Fatalf("write log file failed: %v", err)
	}

	var output bytes.Buffer
	if err := copyLogTail(&output, logFilePath, 3); err != nil {
		t.Fatalf("copyLogTail() error = %v", err)
	}

	if got := output.String(); got != want.String() {
		t.Fatalf("copyLogTail() = %q, want %q", got, want.String())
	}
}

func TestRunLogsRejectsNegativeLineCount(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := runLogs(&output, -1)
	if err == nil {
		t.Fatal("runLogs() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "line count must be >= 0") {
		t.Fatalf("runLogs() error = %v, want line count error", err)
	}
}
