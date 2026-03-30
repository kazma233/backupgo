package notice

import "testing"

func TestTaskReportTracksFirstErrorAndUploadResults(t *testing.T) {
	report := NewTaskReport("task-1")
	report.Reset()
	report.SetCompressedSize(2048)
	report.AddUploadSuccess("archive", "demo.zip")
	report.AddUploadFailure("archive", "demo-2.zip", "network error")
	report.MarkError("上传失败")
	report.EnsureFailed("备份失败")
	report.Finish()

	snapshot := report.Snapshot()
	if snapshot.TaskID != "task-1" {
		t.Fatalf("expected task id task-1, got %q", snapshot.TaskID)
	}
	if !snapshot.HasErrors {
		t.Fatal("expected report to be marked as failed")
	}
	if snapshot.ErrorCount != 1 {
		t.Fatalf("expected error count 1, got %d", snapshot.ErrorCount)
	}
	if snapshot.CompressedSize != "2.0 KB" {
		t.Fatalf("expected compressed size 2.0 KB, got %q", snapshot.CompressedSize)
	}
	if snapshot.FirstError != "上传失败" {
		t.Fatalf("expected first error 上传失败, got %q", snapshot.FirstError)
	}
	if len(snapshot.Uploads) != 2 {
		t.Fatalf("expected 2 uploads, got %d", len(snapshot.Uploads))
	}
	if snapshot.Uploads[0].ObjectPath() != "archive/demo.zip" {
		t.Fatalf("unexpected success upload path: %s", snapshot.Uploads[0].ObjectPath())
	}
	if snapshot.Uploads[1].Reason != "network error" {
		t.Fatalf("unexpected failed upload reason: %s", snapshot.Uploads[1].Reason)
	}
	if snapshot.Duration < 0 {
		t.Fatalf("unexpected negative duration: %s", snapshot.Duration)
	}
}

func TestTaskReportResetClearsState(t *testing.T) {
	report := NewTaskReport("task-1")
	report.MarkError("上传失败")
	report.AddUploadFailure("archive", "demo.zip", "network error")
	report.Reset()

	snapshot := report.Snapshot()
	if snapshot.HasErrors {
		t.Fatal("expected reset report to clear errors")
	}
	if snapshot.ErrorCount != 0 {
		t.Fatalf("expected reset error count 0, got %d", snapshot.ErrorCount)
	}
	if len(snapshot.Uploads) != 0 {
		t.Fatalf("expected reset uploads to be empty, got %d", len(snapshot.Uploads))
	}
	if snapshot.FirstError != "" {
		t.Fatalf("expected reset first error to be empty, got %q", snapshot.FirstError)
	}
}
