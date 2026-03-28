package notice

import (
	"fmt"
	"time"
)

type UploadSummary struct {
	Bucket string
	Key    string
}

type TaskSummary struct {
	TaskID         string
	Duration       time.Duration
	HasErrors      bool
	ErrorCount     int
	CompressedSize string
	Uploads        []UploadSummary
	FirstError     string
}

type TaskLogger struct {
	taskID         string
	startTime      time.Time
	lastEventTime  time.Time
	lines          []string
	hasErrors      bool
	errorCount     int
	firstError     string
	compressedSize string
	uploads        []UploadSummary
}

func NewTaskLogger(taskID string) *TaskLogger {
	logger := &TaskLogger{taskID: taskID}
	logger.StartNewTask()
	return logger
}

func (tl *TaskLogger) LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	tl.appendLine(fmt.Sprintf("[%s] %s", tl.taskID, message))
}

func (tl *TaskLogger) LogError(err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	tl.hasErrors = true
	tl.errorCount++
	if tl.firstError == "" {
		tl.firstError = message
	}
	tl.appendLine(fmt.Sprintf("[%s] %s: %v", tl.taskID, message, err))
}

func (tl *TaskLogger) LogProgress(filePath string, processed, total int64, percentage float64) {
	tl.appendLine(fmt.Sprintf("[%s] 进度: %s - %s / %s (%.1f%%)",
		tl.taskID, filePath, FormatBytes(processed), FormatBytes(total), percentage))
}

func (tl *TaskLogger) LogCompressed(total int64) {
	size := FormatBytes(total)
	tl.compressedSize = size
	tl.appendLine(fmt.Sprintf("[%s] 压缩完成，总大小: %s", tl.taskID, size))
}

func (tl *TaskLogger) LogUpload(bucket string, key string) {
	tl.uploads = append(tl.uploads, UploadSummary{
		Bucket: bucket,
		Key:    key,
	})
	tl.appendLine(fmt.Sprintf("[%s] 上传完成: %s", tl.taskID, key))
}

func (tl *TaskLogger) StartStage(stageName string) {
	tl.appendLine(fmt.Sprintf("【%s】%s 开始", tl.taskID, stageName))
}

func (tl *TaskLogger) FinishStage(stageName string) {
	tl.appendLine(fmt.Sprintf("【%s】%s 完成", tl.taskID, stageName))
}

func (tl *TaskLogger) FailStage(stageName string, err error) {
	tl.hasErrors = true
	if tl.firstError == "" {
		tl.firstError = fmt.Sprintf("失败: %s", stageName)
	}
	tl.appendLine(fmt.Sprintf("[%s] %s 失败: %v", tl.taskID, stageName, err))
}

func (tl *TaskLogger) Lines() []string {
	lines := make([]string, len(tl.lines))
	copy(lines, tl.lines)
	return lines
}

func (tl *TaskLogger) Summary() TaskSummary {
	uploads := make([]UploadSummary, len(tl.uploads))
	copy(uploads, tl.uploads)

	duration := time.Duration(0)
	if !tl.lastEventTime.IsZero() {
		duration = tl.lastEventTime.Sub(tl.startTime)
	}

	return TaskSummary{
		TaskID:         tl.taskID,
		Duration:       duration,
		HasErrors:      tl.hasErrors,
		ErrorCount:     tl.errorCount,
		CompressedSize: tl.compressedSize,
		Uploads:        uploads,
		FirstError:     tl.firstError,
	}
}

func (tl *TaskLogger) StartNewTask() {
	now := time.Now()
	tl.startTime = now
	tl.lastEventTime = now
	tl.lines = make([]string, 0)
	tl.hasErrors = false
	tl.errorCount = 0
	tl.firstError = ""
	tl.compressedSize = ""
	tl.uploads = make([]UploadSummary, 0)
}

func (tl *TaskLogger) appendLine(line string) {
	tl.lines = append(tl.lines, line)
	tl.lastEventTime = time.Now()
	fmt.Println(line)
}
