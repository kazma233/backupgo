package notice

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTaskLoggerAddsStructuredEntries(t *testing.T) {
	logger := NewTaskLogger("task-1")
	logger.StartNewTask()
	logger.LogCompressed(2048)
	logger.LogUpload("archive", "demo.zip")

	entries := logger.GetEntries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].EntryType != EntryTypeCompressed {
		t.Fatalf("expected first entry type %q, got %q", EntryTypeCompressed, entries[0].EntryType)
	}
	if entries[0].CompressedSize != "2.0 KB" {
		t.Fatalf("expected compressed size 2.0 KB, got %q", entries[0].CompressedSize)
	}

	if entries[1].EntryType != EntryTypeUpload {
		t.Fatalf("expected second entry type %q, got %q", EntryTypeUpload, entries[1].EntryType)
	}
	if entries[1].UploadBucket != "archive" || entries[1].UploadKey != "demo.zip" {
		t.Fatalf("unexpected upload entry: %+v", entries[1])
	}
}

func TestBuildTaskSummaryUsesStructuredEntries(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	entries := []LogEntry{
		{EntryType: EntryTypeStep, Timestamp: start.Add(time.Second), StepName: "备份", StepStatus: StepStatusStart, Message: "开始: 备份"},
		{EntryType: EntryTypeCompressed, Timestamp: start.Add(2 * time.Second), Message: "压缩完成，总大小: 12.0 MB", CompressedSize: "12.0 MB"},
		{EntryType: EntryTypeUpload, Timestamp: start.Add(3 * time.Second), Message: "上传完成: demo.zip", UploadBucket: "OSS", UploadKey: "demo.zip"},
		{EntryType: EntryTypeError, Timestamp: start.Add(4 * time.Second), Message: "上传失败", Error: errors.New("boom")},
	}

	summary := buildTaskSummary("task-1", start, entries)
	if summary.taskID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", summary.taskID)
	}
	if summary.statusText != "失败" || summary.statusIcon != "❌" {
		t.Fatalf("expected failed status, got %q %q", summary.statusIcon, summary.statusText)
	}
	if summary.duration != 4*time.Second {
		t.Fatalf("expected duration 4s, got %s", summary.duration)
	}
	if summary.stepCount != 1 {
		t.Fatalf("expected step count 1, got %d", summary.stepCount)
	}
	if summary.errorCount != 1 {
		t.Fatalf("expected error count 1, got %d", summary.errorCount)
	}
	if summary.compressedSize != "12.0 MB" {
		t.Fatalf("expected compressed size 12.0 MB, got %q", summary.compressedSize)
	}
	if len(summary.uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(summary.uploads))
	}
	if summary.uploads[0].bucket != "OSS" || summary.uploads[0].key != "demo.zip" {
		t.Fatalf("unexpected upload info: %+v", summary.uploads[0])
	}
	if summary.firstError != "上传失败" {
		t.Fatalf("expected first error 上传失败, got %q", summary.firstError)
	}
}

func TestFormatterRendersPlainAndHTML(t *testing.T) {
	summary := taskSummary{
		taskID:         "task-1",
		statusIcon:     "✅",
		statusText:     "成功",
		duration:       2*time.Minute + 3*time.Second,
		stepCount:      4,
		errorCount:     0,
		compressedSize: "10.0 MB",
		uploads: []uploadInfo{
			{bucket: "OSS", key: "demo.zip"},
		},
	}

	plain := newFormatter(FormatTypePlain).FormatSummary(summary)
	for _, want := range []string{
		"📦 备份任务: task-1",
		"✅ 状态: 成功",
		"⏱️ 耗时: 2分3秒",
		"📦 10.0 MB",
		"☁️ 上传至: OSS/demo.zip",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("plain output missing %q: %s", want, plain)
		}
	}

	html := newFormatter(FormatTypeHTML).FormatSummary(summary)
	for _, want := range []string{
		"<b>📦 备份任务:</b> <code>task-1</code>",
		"✅ <b>状态:</b> 成功",
		"⏱️ <b>耗时:</b> 2分3秒",
		"📦 <b>压缩:</b> 10.0 MB",
		"☁️ <b>上传至:</b> <code>OSS/demo.zip</code>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html output missing %q: %s", want, html)
		}
	}
}
