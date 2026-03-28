package notice

import (
	"strings"
	"testing"
	"time"
)

func TestTaskLoggerTracksSummaryAndLines(t *testing.T) {
	logger := NewTaskLogger("task-1")
	logger.StartNewTask()
	logger.StartStage("备份")
	logger.LogCompressed(2048)
	logger.LogUpload("archive", "demo.zip")
	logger.LogError(assertError("boom"), "上传失败")
	logger.FinishStage("备份")

	summary := logger.Summary()
	if summary.TaskID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", summary.TaskID)
	}
	if !summary.HasErrors {
		t.Fatal("expected summary to be marked as failed")
	}
	if summary.ErrorCount != 1 {
		t.Fatalf("expected error count 1, got %d", summary.ErrorCount)
	}
	if summary.CompressedSize != "2.0 KB" {
		t.Fatalf("expected compressed size 2.0 KB, got %q", summary.CompressedSize)
	}
	if len(summary.Uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(summary.Uploads))
	}
	if summary.Uploads[0].Bucket != "archive" || summary.Uploads[0].Key != "demo.zip" {
		t.Fatalf("unexpected upload summary: %+v", summary.Uploads[0])
	}
	if summary.FirstError != "上传失败" {
		t.Fatalf("expected first error 上传失败, got %q", summary.FirstError)
	}

	lines := logger.Lines()
	if len(lines) != 5 {
		t.Fatalf("expected 5 log lines, got %d", len(lines))
	}
	for _, want := range []string{
		"【task-1】备份 开始",
		"[task-1] 压缩完成，总大小: 2.0 KB",
		"[task-1] 上传完成: demo.zip",
		"[task-1] 上传失败: boom",
		"【task-1】备份 完成",
	} {
		if !containsLine(lines, want) {
			t.Fatalf("expected lines to contain %q, got %#v", want, lines)
		}
	}
}

func TestTaskLoggerFailStageSetsFailureWithoutDetailedErrorLog(t *testing.T) {
	logger := NewTaskLogger("task-1")
	logger.StartNewTask()
	logger.StartStage("上传到OSS")
	logger.FailStage("上传到OSS", assertError("network"))

	summary := logger.Summary()
	if !summary.HasErrors {
		t.Fatal("expected failed summary")
	}
	if summary.FirstError != "失败: 上传到OSS" {
		t.Fatalf("expected first error to fall back to stage failure, got %q", summary.FirstError)
	}
}

func TestFormatterRendersPlainAndHTML(t *testing.T) {
	summary := TaskSummary{
		TaskID:         "task-1",
		Duration:       2*time.Minute + 3*time.Second,
		CompressedSize: "10.0 MB",
		Uploads: []UploadSummary{
			{Bucket: "OSS", Key: "demo.zip"},
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
	if strings.Contains(plain, "统计") {
		t.Fatalf("plain output should not contain step statistics: %s", plain)
	}

	html := newFormatter(FormatTypeHTML).FormatSummary(summary)
	for _, want := range []string{
		"<div><b>📦 备份任务:</b> <code>task-1</code></div>",
		"<div>✅ <b>状态:</b> 成功</div>",
		"<div>⏱️ <b>耗时:</b> 2分3秒</div>",
		"<div>📦 <b>压缩:</b> 10.0 MB</div>",
		"<div>☁️ <b>上传至:</b> <code>OSS/demo.zip</code></div>",
		"<div><br/></div>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html output missing %q: %s", want, html)
		}
	}
}

func TestFormatterEscapesHTMLContent(t *testing.T) {
	summary := TaskSummary{
		TaskID:         `task<&>`,
		Duration:       5 * time.Second,
		HasErrors:      true,
		CompressedSize: `10<&> MB`,
		Uploads: []UploadSummary{
			{Bucket: `OSS&1`, Key: `demo<zip>`},
		},
		FirstError: `bad <error> & fail`,
	}

	html := newFormatter(FormatTypeHTML).FormatSummary(summary)
	for _, want := range []string{
		"<code>task&lt;&amp;&gt;</code>",
		"<div>❌ <b>状态:</b> 失败</div>",
		"<div>📦 <b>压缩:</b> 10&lt;&amp;&gt; MB</div>",
		"<div>☁️ <b>上传至:</b> <code>OSS&amp;1/demo&lt;zip&gt;</code></div>",
		"<div>❌ <b>错误:</b> <code>bad &lt;error&gt; &amp; fail</code></div>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html output missing escaped content %q: %s", want, html)
		}
	}

	for _, raw := range []string{
		`task<&>`,
		`10<&> MB`,
		`OSS&1/demo<zip>`,
		`bad <error> & fail`,
	} {
		if strings.Contains(html, raw) {
			t.Fatalf("html output should not contain raw content %q: %s", raw, html)
		}
	}
}

func containsLine(lines []string, want string) bool {
	for _, line := range lines {
		if line == want {
			return true
		}
	}
	return false
}

func assertError(message string) error {
	return testError(message)
}

type testError string

func (e testError) Error() string {
	return string(e)
}
