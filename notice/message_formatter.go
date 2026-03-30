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

func (f formatter) FormatReport(report TaskReport) string {
	var builder strings.Builder

	switch f.formatType {
	case FormatTypeMarkdown:
		renderMarkdown(&builder, report)
	case FormatTypeHTML:
		renderHTML(&builder, report)
	default:
		renderPlain(&builder, report)
	}

	return builder.String()
}

func renderPlain(builder *strings.Builder, report TaskReport) {
	writeLine(builder, "📦 备份任务: %s", report.TaskID)
	writeLine(builder, "%s 状态: %s", statusIcon(report.HasErrors), statusText(report.HasErrors))
	writeLine(builder, "⏱️ 耗时: %s", FormatDuration(report.Duration))
	writeSeparator(builder)

	if report.CompressedSize != "" {
		writeLine(builder, "📦 %s", report.CompressedSize)
	}

	for _, upload := range report.Uploads {
		writePlainUpload(builder, upload)
	}

	if report.FirstError != "" {
		writeLine(builder, "❌ 错误: %s", report.FirstError)
	}
}

func renderMarkdown(builder *strings.Builder, report TaskReport) {
	writeLine(builder, "📦 **备份任务**: `%s`", report.TaskID)
	writeLine(builder, "%s **状态**: %s", statusIcon(report.HasErrors), statusText(report.HasErrors))
	writeLine(builder, "⏱️ **耗时**: %s", FormatDuration(report.Duration))
	writeLine(builder, "")
	writeLine(builder, "---")
	writeLine(builder, "")

	if report.CompressedSize != "" {
		writeLine(builder, "📦 **压缩**: %s", report.CompressedSize)
	}

	for _, upload := range report.Uploads {
		writeMarkdownUpload(builder, upload)
	}

	if report.FirstError != "" {
		writeLine(builder, "")
		writeLine(builder, "❌ **错误**: `%s`", report.FirstError)
	}
}

func renderHTML(builder *strings.Builder, report TaskReport) {
	writeHTMLBlock(builder, "<b>📦 备份任务:</b> <code>%s</code>", escapeHTML(report.TaskID))
	writeHTMLBlock(builder, "%s <b>状态:</b> %s", statusIcon(report.HasErrors), escapeHTML(statusText(report.HasErrors)))
	writeHTMLBlock(builder, "⏱️ <b>耗时:</b> %s", escapeHTML(FormatDuration(report.Duration)))
	writeHTMLSpacer(builder)

	if report.CompressedSize != "" {
		writeHTMLBlock(builder, "📦 <b>压缩:</b> %s", escapeHTML(report.CompressedSize))
	}

	for _, upload := range report.Uploads {
		writeHTMLUpload(builder, upload)
	}

	if report.FirstError != "" {
		writeHTMLSpacer(builder)
		writeHTMLBlock(builder, "❌ <b>错误:</b> <code>%s</code>", escapeHTML(report.FirstError))
	}
}

func writeLine(builder *strings.Builder, format string, args ...interface{}) {
	fmt.Fprintf(builder, format+"\n", args...)
}

func writePlainUpload(builder *strings.Builder, upload UploadReport) {
	path := upload.ObjectPath()
	if upload.Status == UploadStatusFailed {
		if upload.Reason != "" {
			writeLine(builder, "☁️ 上传失败: %s (%s)", path, upload.Reason)
			return
		}
		writeLine(builder, "☁️ 上传失败: %s", path)
		return
	}

	writeLine(builder, "☁️ 对象路径: %s", path)
}

func writeMarkdownUpload(builder *strings.Builder, upload UploadReport) {
	path := upload.ObjectPath()
	if upload.Status == UploadStatusFailed {
		if upload.Reason != "" {
			writeLine(builder, "☁️ **上传失败**: `%s` (`%s`)", path, upload.Reason)
			return
		}
		writeLine(builder, "☁️ **上传失败**: `%s`", path)
		return
	}

	writeLine(builder, "☁️ **对象路径**: `%s`", path)
}

func writeHTMLUpload(builder *strings.Builder, upload UploadReport) {
	path := escapeHTML(upload.ObjectPath())
	if upload.Status == UploadStatusFailed {
		if upload.Reason != "" {
			writeHTMLBlock(builder, "☁️ <b>上传失败:</b> <code>%s</code> (<code>%s</code>)", path, escapeHTML(upload.Reason))
			return
		}
		writeHTMLBlock(builder, "☁️ <b>上传失败:</b> <code>%s</code>", path)
		return
	}

	writeHTMLBlock(builder, "☁️ <b>对象路径:</b> <code>%s</code>", path)
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
