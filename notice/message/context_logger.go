package message

import (
	"fmt"
	"time"
)

// === EntryType 日志条目类型 ===

// EntryType 定义日志条目的类型
type EntryType string

const (
	EntryTypeStep     EntryType = "step"     // 步骤开始/结束
	EntryTypeProgress EntryType = "progress" // 进度信息
	EntryTypeInfo     EntryType = "info"     // 一般信息
	EntryTypeError    EntryType = "error"    // 错误信息
)

// StepStatus 定义步骤的状态
type StepStatus string

const (
	StepStatusStart   StepStatus = "start"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
)

// === LogEntry 日志条目 ===

// LogEntry 结构化的日志条目
type LogEntry struct {
	EntryType EntryType
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

// === ConsoleLogger 控制台日志输出 ===

// ConsoleLogger 负责将日志输出到控制台
type ConsoleLogger struct {
	taskID string
}

// NewConsoleLogger 创建新的控制台日志记录器
func NewConsoleLogger(taskID string) *ConsoleLogger {
	return &ConsoleLogger{taskID: taskID}
}

// Log 输出普通日志
func (cl *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", cl.taskID, message)
}

// LogStep 输出步骤日志
func (cl *ConsoleLogger) LogStep(stepName string, status string) {
	fmt.Printf("【%s】%s %s\n", cl.taskID, stepName, status)
}

// LogProgress 输出进度日志
func (cl *ConsoleLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	fmt.Printf("[%s] 进度: %s - %s / %s (%.1f%%)\n",
		cl.taskID, filePath, FormatBytes(processed), FormatBytes(total), percentage)
}

// === TaskLogger 任务日志记录器 ===

// TaskLogger 负责任务日志的收集和管理
type TaskLogger struct {
	taskID        string
	startTime     time.Time
	entries       []LogEntry
	stepStack     []string
	consoleLogger *ConsoleLogger
}

// NewTaskLogger 创建新的任务日志记录器
func NewTaskLogger(taskID string) *TaskLogger {
	return &TaskLogger{
		taskID:        taskID,
		startTime:     time.Now(),
		entries:       make([]LogEntry, 0),
		stepStack:     make([]string, 0),
		consoleLogger: NewConsoleLogger(taskID),
	}
}

// === 日志记录方法 ===

// LogInfo 记录一般信息
func (tl *TaskLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		EntryType: EntryTypeInfo,
		Timestamp: time.Now(),
		Message:   message,
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.Log(message)
}

// LogError 记录错误信息
func (tl *TaskLogger) LogError(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	entry := LogEntry{
		EntryType: EntryTypeError,
		Timestamp: time.Now(),
		Message:   message,
		Error:     err,
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.Log(fmt.Sprintf("%s: %v", message, err))
}

// LogProgress 记录进度信息
func (tl *TaskLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	entry := LogEntry{
		EntryType:  EntryTypeProgress,
		Timestamp:  time.Now(),
		FilePath:   filePath,
		Processed:  processed,
		Total:      total,
		Percentage: percentage,
		Message:    fmt.Sprintf("进度: %s (%.1f%%)", filePath, percentage),
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.LogProgress(filePath, processed, total, percentage)
}

// GetEntries 返回所有日志条目
func (tl *TaskLogger) GetEntries() []LogEntry {
	return tl.entries
}

// GetStartTime 返回任务开始时间
func (tl *TaskLogger) GetStartTime() time.Time {
	return tl.startTime
}

// StartNewTask 开始新的任务，重置所有状态
func (tl *TaskLogger) StartNewTask() {
	tl.startTime = time.Now()
	tl.entries = make([]LogEntry, 0)
	tl.stepStack = make([]string, 0)
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
		EntryType:  EntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusStart,
		Message:    fmt.Sprintf("开始: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.LogStep(stepName, "开始")
}

// stepSuccess 记录步骤成功并出栈
func (tl *TaskLogger) stepSuccess(stepName string) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		EntryType:  EntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusSuccess,
		Message:    fmt.Sprintf("完成: %s", stepName),
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.LogStep(stepName, "完成")
}

// stepFailed 记录步骤失败并出栈
func (tl *TaskLogger) stepFailed(stepName string, err error) {
	// 出栈
	if len(tl.stepStack) > 0 {
		tl.stepStack = tl.stepStack[:len(tl.stepStack)-1]
	}

	entry := LogEntry{
		EntryType:  EntryTypeStep,
		Timestamp:  time.Now(),
		StepName:   stepName,
		StepStatus: StepStatusFailed,
		Message:    fmt.Sprintf("失败: %s", stepName),
		Error:      err,
	}
	tl.entries = append(tl.entries, entry)
	tl.consoleLogger.Log(fmt.Sprintf("%s 失败: %v", stepName, err))
}
