package message

import (
	"fmt"
	"strings"
	"time"
)

// === MessageFormatter ===

// MessageFormatter 定义消息格式化器接口
type MessageFormatter interface {
	Format(taskID string, startTime time.Time, entries []LogEntry) string
}

// SimpleTextFormatter 简化的文本格式化器，只显示核心摘要信息
type SimpleTextFormatter struct {
	builder strings.Builder
}

// NewSimpleTextFormatter 创建新的简化文本格式化器
func NewSimpleTextFormatter() *SimpleTextFormatter {
	return &SimpleTextFormatter{}
}

// Format 将日志条目格式化为简化的纯文本消息
func (f *SimpleTextFormatter) Format(taskID string, startTime time.Time, entries []LogEntry) string {
	f.builder.Reset()

	status := "成功"
	if hasErrors(entries) {
		status = "失败"
	}
	fmt.Fprintf(&f.builder, "备份任务: %s - 状态: %s\n", taskID, status)

	endTime := startTime
	if len(entries) > 0 {
		endTime = entries[len(entries)-1].Timestamp
	}
	duration := endTime.Sub(startTime)
	fmt.Fprintf(&f.builder, "耗时: %s\n", FormatDuration(duration))

	// 收集关键信息
	var compressedSize string
	var uploadInfos []string

	for _, entry := range entries {
		switch entry.EntryType {
		case EntryTypeInfo:
			// 提取压缩完成信息
			if strings.Contains(entry.Message, "压缩完成") {
				compressedSize = entry.Message
			}
			// 提取上传信息
			if strings.Contains(entry.Message, "bucket") {
				uploadInfos = append(uploadInfos, entry.Message)
			}
		}
	}

	// 输出压缩信息
	if compressedSize != "" {
		fmt.Fprintf(&f.builder, "%s\n", compressedSize)
	}

	// 输出上传信息
	for _, info := range uploadInfos {
		fmt.Fprintf(&f.builder, "上传至: %s\n", info)
	}

	// 输出错误信息
	if firstError := f.getFirstError(entries); firstError != "" {
		fmt.Fprintf(&f.builder, "错误: %s\n", firstError)
	}

	return f.builder.String()
}

// getFirstError 获取第一个错误信息
func (f *SimpleTextFormatter) getFirstError(entries []LogEntry) string {
	for _, entry := range entries {
		if entry.EntryType == EntryTypeError {
			return entry.Message
		}
		if entry.EntryType == EntryTypeStep && entry.StepStatus == StepStatusFailed {
			return entry.Message
		}
	}
	return ""
}

// hasErrors 检查日志条目中是否有错误
func hasErrors(entries []LogEntry) bool {
	for _, entry := range entries {
		if entry.EntryType == EntryTypeError {
			return true
		}
		if entry.EntryType == EntryTypeStep && entry.StepStatus == StepStatusFailed {
			return true
		}
	}
	return false
}

// === tools ===

// FormatBytes 将字节数转换为人类可读的格式
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	} else {
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	}
}

// FormatDuration 将时间间隔转换为易读格式
func FormatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

// FormatTimestamp 格式化时间戳
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
