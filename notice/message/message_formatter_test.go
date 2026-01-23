package message

import (
	"testing"
	"time"
)

func TestSimpleTextFormatter_Success(t *testing.T) {
	formatter := NewSimpleTextFormatter()
	startTime := time.Now()

	entries := []LogEntry{
		{
			Type:      LogEntryTypeInfo,
			Timestamp: startTime.Add(time.Second * 1),
			Message:   "压缩完成，总大小: 100.5 MB",
		},
		{
			Type:      LogEntryTypeInfo,
			Timestamp: startTime.Add(time.Second * 2),
			Message:   "上传到 bucket: mybucket",
		},
	}

	result := formatter.Format("test_task", startTime, entries)

	t.Logf("成功场景输出:\n%s", result)

	expectedContains := []string{
		"备份任务: test_task - 状态: 成功",
		"耗时:",
		"压缩完成，总大小: 100.5 MB",
		"上传至: 上传到 bucket: mybucket",
	}

	for _, expected := range expectedContains {
		if !contains(result, expected) {
			t.Errorf("期望包含 '%s'，但输出为:\n%s", expected, result)
		}
	}

	if contains(result, "错误") {
		t.Errorf("成功场景不应包含错误信息，输出为:\n%s", result)
	}
}

func TestSimpleTextFormatter_Failure(t *testing.T) {
	formatter := NewSimpleTextFormatter()
	startTime := time.Now()

	entries := []LogEntry{
		{
			Type:      LogEntryTypeError,
			Timestamp: startTime.Add(time.Second * 1),
			Message:   "压缩失败",
			Error:     nil,
		},
	}

	result := formatter.Format("test_task", startTime, entries)

	t.Logf("失败场景输出:\n%s", result)

	expectedContains := []string{
		"备份任务: test_task - 状态: 失败",
		"耗时:",
		"错误: 压缩失败",
	}

	for _, expected := range expectedContains {
		if !contains(result, expected) {
			t.Errorf("期望包含 '%s'，但输出为:\n%s", expected, result)
		}
	}
}

func TestSimpleTextFormatter_Duration(t *testing.T) {
	formatter := NewSimpleTextFormatter()
	startTime := time.Now()

	entries := []LogEntry{
		{
			Type:      LogEntryTypeInfo,
			Timestamp: startTime.Add(time.Minute*1 + time.Second*30),
			Message:   "压缩完成",
		},
	}

	result := formatter.Format("test_task", startTime, entries)

	if !contains(result, "1分30秒") {
		t.Errorf("期望包含 '1分30秒'，输出为:\n%s", result)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
