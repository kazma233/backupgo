package notice

import (
	"fmt"
	"time"
)

type UploadStatus string

const (
	UploadStatusSuccess UploadStatus = "success"
	UploadStatusFailed  UploadStatus = "failed"
)

type UploadReport struct {
	Bucket string
	Key    string
	Status UploadStatus
	Reason string
}

func (u UploadReport) ObjectPath() string {
	switch {
	case u.Bucket != "" && u.Key != "":
		return fmt.Sprintf("%s/%s", u.Bucket, u.Key)
	case u.Key != "":
		return u.Key
	case u.Bucket != "":
		return u.Bucket
	default:
		return "未生成对象路径"
	}
}

type TaskReport struct {
	TaskID         string
	Duration       time.Duration
	HasErrors      bool
	ErrorCount     int
	CompressedSize string
	Uploads        []UploadReport
	FirstError     string

	startedAt time.Time
}

func NewTaskReport(taskID string) *TaskReport {
	report := &TaskReport{TaskID: taskID}
	report.Reset()
	return report
}

func (r *TaskReport) Reset() {
	r.Duration = 0
	r.HasErrors = false
	r.ErrorCount = 0
	r.CompressedSize = ""
	r.Uploads = make([]UploadReport, 0)
	r.FirstError = ""
	r.startedAt = time.Now()
}

func (r *TaskReport) Finish() {
	if r.startedAt.IsZero() {
		return
	}
	r.Duration = time.Since(r.startedAt)
}

func (r *TaskReport) MarkError(message string) {
	r.HasErrors = true
	r.ErrorCount++
	if r.FirstError == "" {
		r.FirstError = message
	}
}

func (r *TaskReport) EnsureFailed(message string) {
	r.HasErrors = true
	if r.FirstError == "" {
		r.FirstError = message
	}
}

func (r *TaskReport) SetCompressedSize(total int64) {
	r.CompressedSize = FormatBytes(total)
}

func (r *TaskReport) AddUploadSuccess(bucket string, key string) {
	r.Uploads = append(r.Uploads, UploadReport{
		Bucket: bucket,
		Key:    key,
		Status: UploadStatusSuccess,
	})
}

func (r *TaskReport) AddUploadFailure(bucket string, key string, reason string) {
	r.Uploads = append(r.Uploads, UploadReport{
		Bucket: bucket,
		Key:    key,
		Status: UploadStatusFailed,
		Reason: reason,
	})
}

func (r *TaskReport) Snapshot() TaskReport {
	uploads := make([]UploadReport, len(r.Uploads))
	copy(uploads, r.Uploads)

	return TaskReport{
		TaskID:         r.TaskID,
		Duration:       r.Duration,
		HasErrors:      r.HasErrors,
		ErrorCount:     r.ErrorCount,
		CompressedSize: r.CompressedSize,
		Uploads:        uploads,
		FirstError:     r.FirstError,
	}
}
