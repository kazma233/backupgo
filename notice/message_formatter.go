package notice

import (
	"fmt"
	"html"
	"strings"
	"time"
)

type FormatType string

const (
	FormatTypePlain    FormatType = "plain"
	FormatTypeMarkdown FormatType = "markdown"
	FormatTypeHTML     FormatType = "html"
)

type formatter struct {
	formatType FormatType
}

func newFormatter(formatType FormatType) formatter {
	return formatter{formatType: formatType}
}

func (f formatter) FormatSummary(summary TaskSummary) string {
	var builder strings.Builder

	switch f.formatType {
	case FormatTypeMarkdown:
		renderMarkdown(&builder, summary)
	case FormatTypeHTML:
		renderHTML(&builder, summary)
	default:
		renderPlain(&builder, summary)
	}

	return builder.String()
}

func renderPlain(builder *strings.Builder, summary TaskSummary) {
	writeLine(builder, "📦 备份任务: %s", summary.TaskID)
	writeLine(builder, "%s 状态: %s", statusIcon(summary.HasErrors), statusText(summary.HasErrors))
	writeLine(builder, "⏱️ 耗时: %s", FormatDuration(summary.Duration))
	writeSeparator(builder)

	if summary.CompressedSize != "" {
		writeLine(builder, "📦 %s", summary.CompressedSize)
	}

	for _, upload := range summary.Uploads {
		writeLine(builder, "☁️ 上传至: %s/%s", upload.Bucket, upload.Key)
	}

	if summary.FirstError != "" {
		writeLine(builder, "❌ 错误: %s", summary.FirstError)
	}
}

func renderMarkdown(builder *strings.Builder, summary TaskSummary) {
	writeLine(builder, "📦 **备份任务**: `%s`", summary.TaskID)
	writeLine(builder, "%s **状态**: %s", statusIcon(summary.HasErrors), statusText(summary.HasErrors))
	writeLine(builder, "⏱️ **耗时**: %s", FormatDuration(summary.Duration))
	writeLine(builder, "")
	writeLine(builder, "---")
	writeLine(builder, "")

	if summary.CompressedSize != "" {
		writeLine(builder, "📦 **压缩**: %s", summary.CompressedSize)
	}

	for _, upload := range summary.Uploads {
		writeLine(builder, "☁️ **上传至**: `%s/%s`", upload.Bucket, upload.Key)
	}

	if summary.FirstError != "" {
		writeLine(builder, "")
		writeLine(builder, "❌ **错误**: `%s`", summary.FirstError)
	}
}

func renderHTML(builder *strings.Builder, summary TaskSummary) {
	writeHTMLBlock(builder, "<b>📦 备份任务:</b> <code>%s</code>", escapeHTML(summary.TaskID))
	writeHTMLBlock(builder, "%s <b>状态:</b> %s", statusIcon(summary.HasErrors), escapeHTML(statusText(summary.HasErrors)))
	writeHTMLBlock(builder, "⏱️ <b>耗时:</b> %s", escapeHTML(FormatDuration(summary.Duration)))
	writeHTMLSpacer(builder)

	if summary.CompressedSize != "" {
		writeHTMLBlock(builder, "📦 <b>压缩:</b> %s", escapeHTML(summary.CompressedSize))
	}

	for _, upload := range summary.Uploads {
		writeHTMLBlock(builder, "☁️ <b>上传至:</b> <code>%s/%s</code>", escapeHTML(upload.Bucket), escapeHTML(upload.Key))
	}

	if summary.FirstError != "" {
		writeHTMLSpacer(builder)
		writeHTMLBlock(builder, "❌ <b>错误:</b> <code>%s</code>", escapeHTML(summary.FirstError))
	}
}

func writeLine(builder *strings.Builder, format string, args ...interface{}) {
	fmt.Fprintf(builder, format+"\n", args...)
}

func writeHTMLBlock(builder *strings.Builder, format string, args ...interface{}) {
	fmt.Fprintf(builder, "<div>%s</div>\n", fmt.Sprintf(format, args...))
}

func writeHTMLSpacer(builder *strings.Builder) {
	builder.WriteString("<div><br/></div>\n")
}

func escapeHTML(value string) string {
	return html.EscapeString(value)
}

func writeSeparator(builder *strings.Builder) {
	writeLine(builder, "━━━━━━━━━━━━━━━━━━━━")
}

func statusIcon(hasErrors bool) string {
	if hasErrors {
		return "❌"
	}
	return "✅"
}

func statusText(hasErrors bool) string {
	if hasErrors {
		return "失败"
	}
	return "成功"
}

func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	}
	if bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
}

func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	}
	return fmt.Sprintf("%d秒", seconds)
}
