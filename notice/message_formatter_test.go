package notice

import (
	"strings"
	"testing"
	"time"
)

func TestFormatterRendersPlainAndHTML(t *testing.T) {
	report := TaskReport{
		TaskID:         "task-1",
		Duration:       2*time.Minute + 3*time.Second,
		CompressedSize: "10.0 MB",
		Uploads: []UploadReport{
			{Bucket: "OSS", Key: "demo.zip", Status: UploadStatusSuccess},
		},
	}

	plain := newFormatter(FormatTypePlain).FormatReport(report)
	for _, want := range []string{
		"📦 备份任务: task-1",
		"✅ 状态: 成功",
		"⏱️ 耗时: 2分3秒",
		"📦 10.0 MB",
		"☁️ 对象路径: OSS/demo.zip",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("plain output missing %q: %s", want, plain)
		}
	}

	html := newFormatter(FormatTypeHTML).FormatReport(report)
	for _, want := range []string{
		"<div><b>📦 备份任务:</b> <code>task-1</code></div>",
		"<div>✅ <b>状态:</b> 成功</div>",
		"<div>⏱️ <b>耗时:</b> 2分3秒</div>",
		"<div>📦 <b>压缩:</b> 10.0 MB</div>",
		"<div>☁️ <b>对象路径:</b> <code>OSS/demo.zip</code></div>",
		"<div><br/></div>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html output missing %q: %s", want, html)
		}
	}
}

func TestFormatterEscapesHTMLContent(t *testing.T) {
	report := TaskReport{
		TaskID:         `task<&>`,
		Duration:       5 * time.Second,
		HasErrors:      true,
		CompressedSize: `10<&> MB`,
		Uploads: []UploadReport{
			{Bucket: `OSS&1`, Key: `demo<zip>`, Status: UploadStatusSuccess},
		},
		FirstError: `bad <error> & fail`,
	}

	html := newFormatter(FormatTypeHTML).FormatReport(report)
	for _, want := range []string{
		"<code>task&lt;&amp;&gt;</code>",
		"<div>❌ <b>状态:</b> 失败</div>",
		"<div>📦 <b>压缩:</b> 10&lt;&amp;&gt; MB</div>",
		"<div>☁️ <b>对象路径:</b> <code>OSS&amp;1/demo&lt;zip&gt;</code></div>",
		"<div>❌ <b>错误:</b> <code>bad &lt;error&gt; &amp; fail</code></div>",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("html output missing escaped content %q: %s", want, html)
		}
	}
}

func TestFormatterRendersFailedUpload(t *testing.T) {
	report := TaskReport{
		TaskID:    "task-1",
		Duration:  5 * time.Second,
		HasErrors: true,
		Uploads: []UploadReport{
			{Bucket: "OSS", Key: "demo.zip", Status: UploadStatusFailed, Reason: "加速上传冷却中"},
		},
		FirstError: "上传失败",
	}

	plain := newFormatter(FormatTypePlain).FormatReport(report)
	if !strings.Contains(plain, "☁️ 上传失败: OSS/demo.zip (加速上传冷却中)") {
		t.Fatalf("plain output missing failed upload: %s", plain)
	}

	html := newFormatter(FormatTypeHTML).FormatReport(report)
	if !strings.Contains(html, "<div>☁️ <b>上传失败:</b> <code>OSS/demo.zip</code> (<code>加速上传冷却中</code>)</div>") {
		t.Fatalf("html output missing failed upload: %s", html)
	}
}
