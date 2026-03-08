package notice

import (
	"fmt"
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

type uploadInfo struct {
	bucket string
	key    string
}

type taskSummary struct {
	taskID         string
	statusIcon     string
	statusText     string
	duration       time.Duration
	stepCount      int
	errorCount     int
	compressedSize string
	uploads        []uploadInfo
	firstError     string
}

type taskStats struct {
	hasErrors      bool
	errorCount     int
	stepCount      int
	compressedSize string
	uploads        []uploadInfo
	firstError     string
}

func buildTaskSummary(taskID string, startTime time.Time, entries []LogEntry) taskSummary {
	stats := buildTaskStats(entries)
	return taskSummary{
		taskID:         taskID,
		statusIcon:     statusIcon(stats.hasErrors),
		statusText:     statusText(stats.hasErrors),
		duration:       calculateDuration(startTime, entries),
		stepCount:      stats.stepCount,
		errorCount:     stats.errorCount,
		compressedSize: stats.compressedSize,
		uploads:        stats.uploads,
		firstError:     stats.firstError,
	}
}

func buildTaskStats(entries []LogEntry) taskStats {
	stats := taskStats{
		uploads: make([]uploadInfo, 0),
	}

	for _, entry := range entries {
		switch entry.EntryType {
		case EntryTypeStep:
			stats.stepCount++
			if entry.StepStatus == StepStatusFailed {
				stats.hasErrors = true
				stats.errorCount++
				if stats.firstError == "" {
					stats.firstError = firstErrorMessage(entry)
				}
			}
		case EntryTypeError:
			stats.hasErrors = true
			stats.errorCount++
			if stats.firstError == "" {
				stats.firstError = firstErrorMessage(entry)
			}
		case EntryTypeCompressed:
			if entry.CompressedSize != "" {
				stats.compressedSize = entry.CompressedSize
			}
		case EntryTypeUpload:
			if entry.UploadKey == "" {
				continue
			}

			bucket := entry.UploadBucket
			if bucket == "" {
				bucket = "OSS"
			}

			stats.uploads = append(stats.uploads, uploadInfo{
				bucket: bucket,
				key:    entry.UploadKey,
			})
		}
	}

	return stats
}

func firstErrorMessage(entry LogEntry) string {
	if entry.Message != "" {
		return entry.Message
	}
	if entry.Error != nil {
		return entry.Error.Error()
	}
	return ""
}

func (f formatter) Format(taskID string, startTime time.Time, entries []LogEntry) string {
	return f.FormatSummary(buildTaskSummary(taskID, startTime, entries))
}

func (f formatter) FormatSummary(summary taskSummary) string {
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

func renderPlain(builder *strings.Builder, summary taskSummary) {
	writeLine(builder, "📦 备份任务: %s", summary.taskID)
	writeLine(builder, "%s 状态: %s", summary.statusIcon, summary.statusText)
	writeLine(builder, "⏱️ 耗时: %s", FormatDuration(summary.duration))
	writeLine(builder, "📊 统计: %d个步骤 | %d个错误", summary.stepCount, summary.errorCount)
	writeSeparator(builder)

	if summary.compressedSize != "" {
		writeLine(builder, "📦 %s", summary.compressedSize)
	}

	for _, upload := range summary.uploads {
		writeLine(builder, "☁️ 上传至: %s/%s", upload.bucket, upload.key)
	}

	if summary.firstError != "" {
		writeLine(builder, "❌ 错误: %s", summary.firstError)
	}
}

func renderMarkdown(builder *strings.Builder, summary taskSummary) {
	writeLine(builder, "📦 **备份任务**: `%s`", summary.taskID)
	writeLine(builder, "%s **状态**: %s", summary.statusIcon, summary.statusText)
	writeLine(builder, "⏱️ **耗时**: %s", FormatDuration(summary.duration))
	writeLine(builder, "📊 **统计**: %d个步骤 | %d个错误", summary.stepCount, summary.errorCount)
	writeLine(builder, "")
	writeLine(builder, "---")
	writeLine(builder, "")

	if summary.compressedSize != "" {
		writeLine(builder, "📦 **压缩**: %s", summary.compressedSize)
	}

	for _, upload := range summary.uploads {
		writeLine(builder, "☁️ **上传至**: `%s/%s`", upload.bucket, upload.key)
	}

	if summary.firstError != "" {
		writeLine(builder, "")
		writeLine(builder, "❌ **错误**: `%s`", summary.firstError)
	}
}

func renderHTML(builder *strings.Builder, summary taskSummary) {
	writeLine(builder, "<b>📦 备份任务:</b> <code>%s</code>", summary.taskID)
	writeLine(builder, "%s <b>状态:</b> %s", summary.statusIcon, summary.statusText)
	writeLine(builder, "⏱️ <b>耗时:</b> %s", FormatDuration(summary.duration))
	writeLine(builder, "📊 <b>统计:</b> %d个步骤 | %d个错误", summary.stepCount, summary.errorCount)
	writeLine(builder, "")

	if summary.compressedSize != "" {
		writeLine(builder, "📦 <b>压缩:</b> %s", summary.compressedSize)
	}

	for _, upload := range summary.uploads {
		writeLine(builder, "☁️ <b>上传至:</b> <code>%s/%s</code>", upload.bucket, upload.key)
	}

	if summary.firstError != "" {
		writeLine(builder, "")
		writeLine(builder, "❌ <b>错误:</b> <code>%s</code>", summary.firstError)
	}
}

func writeLine(builder *strings.Builder, format string, args ...interface{}) {
	fmt.Fprintf(builder, format+"\n", args...)
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

func calculateDuration(startTime time.Time, entries []LogEntry) time.Duration {
	endTime := startTime
	if len(entries) > 0 {
		endTime = entries[len(entries)-1].Timestamp
	}
	return endTime.Sub(startTime)
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
