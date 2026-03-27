package notice

import (
	"fmt"
	"time"
)

type EntryType string

const (
	EntryTypeStep       EntryType = "step"
	EntryTypeProgress   EntryType = "progress"
	EntryTypeInfo       EntryType = "info"
	EntryTypeError      EntryType = "error"
	EntryTypeCompressed EntryType = "compressed"
	EntryTypeUpload     EntryType = "upload"
)

type StepStatus string

const (
	StepStatusStart   StepStatus = "start"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
)

type LogEntry struct {
	EntryType EntryType
	Timestamp time.Time
	Message   string

	StepName   string
	StepStatus StepStatus

	FilePath   string
	Processed  int64
	Total      int64
	Percentage float64

	CompressedSize string
	UploadBucket   string
	UploadKey      string

	Error error
}

type ConsoleLogger struct {
	taskID string
}

func NewConsoleLogger(taskID string) *ConsoleLogger {
	return &ConsoleLogger{taskID: taskID}
}

func (cl *ConsoleLogger) Log(message string) {
	fmt.Printf("[%s] %s\n", cl.taskID, message)
}

func (cl *ConsoleLogger) LogStep(stepName string, status string) {
	fmt.Printf("【%s】%s %s\n", cl.taskID, stepName, status)
}

func (cl *ConsoleLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	fmt.Printf("[%s] 进度: %s - %s / %s (%.1f%%)\n",
		cl.taskID, filePath, FormatBytes(processed), FormatBytes(total), percentage)
}

type TaskLogger struct {
	startTime     time.Time
	entries       []LogEntry
	consoleLogger *ConsoleLogger
}

func NewTaskLogger(taskID string) *TaskLogger {
	return &TaskLogger{
		startTime:     time.Now(),
		entries:       make([]LogEntry, 0),
		consoleLogger: NewConsoleLogger(taskID),
	}
}

func (tl *TaskLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType: EntryTypeInfo,
		Timestamp: timestamp,
		Message:   message,
	})
	tl.consoleLogger.Log(message)
}

func (tl *TaskLogger) LogError(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType: EntryTypeError,
		Timestamp: timestamp,
		Message:   message,
		Error:     err,
	})
	tl.consoleLogger.Log(fmt.Sprintf("%s: %v", message, err))
}

func (tl *TaskLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:  EntryTypeProgress,
		Timestamp:  timestamp,
		FilePath:   filePath,
		Processed:  processed,
		Total:      total,
		Percentage: percentage,
		Message:    fmt.Sprintf("进度: %s (%.1f%%)", filePath, percentage),
	})
	tl.consoleLogger.LogProgress(filePath, processed, total, percentage)
}

func (tl *TaskLogger) LogCompressed(total int64) {
	size := FormatBytes(total)
	message := fmt.Sprintf("压缩完成，总大小: %s", size)
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:      EntryTypeCompressed,
		Timestamp:      timestamp,
		Message:        message,
		CompressedSize: size,
	})
	tl.consoleLogger.Log(message)
}

func (tl *TaskLogger) LogUpload(bucket string, key string) {
	message := fmt.Sprintf("上传完成: %s", key)
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:    EntryTypeUpload,
		Timestamp:    timestamp,
		Message:      message,
		UploadBucket: bucket,
		UploadKey:    key,
	})
	tl.consoleLogger.Log(message)
}

func (tl *TaskLogger) GetEntries() []LogEntry {
	entries := make([]LogEntry, len(tl.entries))
	copy(entries, tl.entries)
	return entries
}

func (tl *TaskLogger) GetStartTime() time.Time {
	return tl.startTime
}

func (tl *TaskLogger) StartNewTask() {
	tl.startTime = time.Now()
	tl.entries = make([]LogEntry, 0)
}

func (tl *TaskLogger) ExecuteStep(stepName string, fn func() error) (err error) {
	tl.stepStart(stepName)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			tl.stepFailed(stepName, err)
			return
		}

		if err != nil {
			tl.stepFailed(stepName, err)
			return
		}

		tl.stepSuccess(stepName)
	}()

	return fn()
}

func (tl *TaskLogger) appendEntry(entry LogEntry) {
	tl.entries = append(tl.entries, entry)
}

func (tl *TaskLogger) stepStart(stepName string) {
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:  EntryTypeStep,
		Timestamp:  timestamp,
		StepName:   stepName,
		StepStatus: StepStatusStart,
		Message:    fmt.Sprintf("开始: %s", stepName),
	})
	tl.consoleLogger.LogStep(stepName, "开始")
}

func (tl *TaskLogger) stepSuccess(stepName string) {
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:  EntryTypeStep,
		Timestamp:  timestamp,
		StepName:   stepName,
		StepStatus: StepStatusSuccess,
		Message:    fmt.Sprintf("完成: %s", stepName),
	})
	tl.consoleLogger.LogStep(stepName, "完成")
}

func (tl *TaskLogger) stepFailed(stepName string, err error) {
	timestamp := time.Now()
	tl.appendEntry(LogEntry{
		EntryType:  EntryTypeStep,
		Timestamp:  timestamp,
		StepName:   stepName,
		StepStatus: StepStatusFailed,
		Message:    fmt.Sprintf("失败: %s", stepName),
		Error:      err,
	})
	tl.consoleLogger.Log(fmt.Sprintf("%s 失败: %v", stepName, err))
}
