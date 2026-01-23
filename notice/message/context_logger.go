package message

import (
	"fmt"
	"log"
	"time"
)

// === LogEntry ===

// LogEntryType 定义日志条目的类型
type LogEntryType string

const (
	LogEntryTypeStep     LogEntryType = "step"     // 步骤开始/结束
	LogEntryTypeProgress LogEntryType = "progress" // 进度信息
	LogEntryTypeInfo     LogEntryType = "info"     // 一般信息
	LogEntryTypeError    LogEntryType = "error"    // 错误信息
)

// StepStatus 定义步骤的状态
type StepStatus string

const (
	StepStatusStart   StepStatus = "start"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
)

// LogEntry 结构化的日志条目
type LogEntry struct {
	Type      LogEntryType
	Timestamp time.Time
	Message   string

	// 步骤相关字段
	StepName   string
	StepStatus StepStatus

	// 进度相关字段
	FilePath   string
	Processed  int64
	Total      int64
	Percentage float64

	// 错误相关字段
	Error error
}

// === TaskLogger ===

// TaskLogger 任务日志记录器，负责日志收集和打印
type TaskLogger struct {
	taskID    string
	startTime time.Time
	entries   []LogEntry // 结构化日志条目
	stepStack []string   // 用于追踪嵌套步骤
}

// NewTaskLogger 创建新的任务日志记录器
func NewTaskLogger(taskID string) *TaskLogger {
	return &TaskLogger{
		taskID:    taskID,
		startTime: time.Now(),
		entries:   make([]LogEntry, 0),
		stepStack: make([]string, 0),
	}
}

// === 结构化日志方法 ===

// LogInfo 记录一般信息
func (tl *TaskLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Type:      LogEntryTypeInfo,
		Timestamp: time.Now(),
		Message:   message,
	}
	tl.entries = append(tl.entries, entry)
	log.Println(message)
}

// LogError 记录错误信息
func (tl *TaskLogger) LogError(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		Type:      LogEntryTypeError,
		Timestamp: time.Now(),
		Message:   message,
		Error:     err,
	}
	tl.entries = append(tl.entries, entry)
	fullMessage := fmt.Sprintf("%s: %v", message, err)
	log.Println(fullMessage)
}

// LogProgress 记录进度信息
func (tl *TaskLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	entry := LogEntry{
		Type:       LogEntryTypeProgress,
		Timestamp:  time.Now(),
		FilePath:   filePath,
		Processed:  processed,
		Total:      total,
		Percentage: percentage,
		Message:    fmt.Sprintf("进度: %s (%.1f%%)", filePath, percentage),
	}
	tl.entries = append(tl.entries, entry)
	message := fmt.Sprintf("进度: %s - %s / %s (%.1f%%)",
		filePath, FormatBytes(processed), FormatBytes(total), percentage)
	log.Println(message)
}

// GetEntries 返回所有日志条目
func (tl *TaskLogger) GetEntries() []LogEntry {
	return tl.entries
}

// GetStartTime 返回任务开始时间
func (tl *TaskLogger) GetStartTime() time.Time {
	return tl.startTime
}

// ExecuteStep 执行一个步骤，自动处理开始、成功和失败状态
// fn 可以返回 error 或者 panic，ExecuteStep 会自动捕获 panic 并转换为 error
func (tl *TaskLogger) ExecuteStep(stepName string, fn func() error) (err error) {
	tl.stepStart(stepName)

	defer func() {
		if r := recover(); r != nil {
			// 捕获 panic 并转换为 error
			err = fmt.Errorf("panic: %v", r)
			tl.stepFailed(stepName, err)
		} else if err != nil {
			// 函数返回了错误
			tl.stepFailed(stepName, err)
		} else {
			// 成功完成
			tl.stepSuccess(stepName)
		}
	}()

	return fn()
}

// stepStart 记录步骤开始并入栈
func (tl *TaskLogger) stepStart(stepName string) {
	tl.stepStack = append(tl.stepStack, stepName)
	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusStart,
		Message:    fmt.Sprintf("开始: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)
	message := fmt.Sprintf("【%s】%s 开始", tl.taskID, stepName)
	log.Println(message)
}

// stepSuccess 记录步骤成功并出栈
func (tl *TaskLogger) stepSuccess(stepName string) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusSuccess,
		Message:    fmt.Sprintf("完成: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)
	message := fmt.Sprintf("【%s】%s 完成", tl.taskID, stepName)
	log.Println(message)
}

// stepFailed 记录步骤失败并出栈
func (tl *TaskLogger) stepFailed(stepName string, err error) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		Type:       LogEntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusFailed,
		Message:    fmt.Sprintf("失败: %s", stepName),
		Error:      err,
	}
	tl.entries = append(tl.entries, entry)
	message := fmt.Sprintf("【%s】%s 失败: %v", tl.taskID, stepName, err)
	log.Println(message)
}
