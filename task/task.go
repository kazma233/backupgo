package task

import (
	"backupgo/config"
	"backupgo/exporter"
	"backupgo/notice"
	"backupgo/oss"
	"backupgo/state"
	"backupgo/utils"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

type TaskHolder struct {
	ID            string
	conf          config.BackupConfig
	ossClient     *oss.OssClient
	noticeManager *notice.NoticeManager
	logger        *slog.Logger
	report        *notice.TaskReport
}

func NewTaskHolder(conf config.BackupConfig, ossClient *oss.OssClient, noticeManager *notice.NoticeManager) *TaskHolder {
	if err := conf.Validate(); err != nil {
		panic(err)
	}

	holder := &TaskHolder{
		ID:            conf.GetID(),
		conf:          conf,
		ossClient:     ossClient,
		noticeManager: noticeManager,
		logger:        slog.Default().With("component", "backup_task", "task_id", conf.GetID()),
		report:        notice.NewTaskReport(conf.GetID()),
	}
	return holder
}

func (c *TaskHolder) BackupTask() {
	c.report.Reset()
	c.logger.Info("backup task started")

	if err := c.backup(); err != nil {
		state.GetState().SetTaskRun(c.ID, "failed")
		c.report.Finish()
		c.sendMessages()
		return
	}

	state.GetState().SetTaskRun(c.ID, "success")

	if err := c.cleanHistory(); err != nil {
		state.GetState().SetTaskRun(c.ID, "failed")
	}

	c.report.Finish()
	c.logger.Info("backup task completed", "status", taskStatus(c.report.HasErrors))
	c.sendMessages()
}

func (c *TaskHolder) cleanHistory() error {
	const stageName = "清理历史文件"
	c.logStageStart(stageName)

	deleted, err := c.ossClient.DeleteObjectsByPredicate(func(key string) bool {
		return utils.IsNeedDeleteFile(c.ID, key)
	})
	if err != nil {
		c.logger.Error("clean history failed", "stage", stageName, "error", err)
		c.report.MarkError("清理历史文件失败")
		return err
	}

	if len(deleted) == 0 {
		c.logger.Info("no historical objects need deletion", "stage", stageName)
		c.logStageFinish(stageName)
		return nil
	}

	c.logger.Info("historical objects deleted", "stage", stageName, "deleted_count", len(deleted), "deleted_keys", deleted)
	c.logStageFinish(stageName)
	return nil
}

func (c *TaskHolder) backup() error {
	const stageName = "备份"
	conf := c.conf

	c.logStageStart(stageName)

	if conf.BeforeCmd != "" {
		if err := c.runCommandStep("执行前置命令", conf.BeforeCmd, "前置命令执行失败"); err != nil {
			c.logStageError(stageName, "backup stage failed", err)
			c.report.EnsureFailed("备份失败")
			return err
		}
	}

	prepared, err := exporter.Prepare(c.ID, conf, c.logger)
	if err != nil {
		c.logger.Error("backup data preparation failed", "stage", stageName, "error", err)
		c.report.MarkError("备份准备失败")
		return err
	}
	defer func() {
		const cleanupStageName = "清理临时文件"

		c.logStageStart(cleanupStageName)
		if cleanupErr := prepared.Cleanup(); cleanupErr != nil {
			c.logger.Error("temporary cleanup failed", "stage", cleanupStageName, "error", cleanupErr)
			c.report.MarkError("清理临时文件失败")
		} else {
			c.logStageFinish(cleanupStageName)
		}

	}()

	c.logger.Info("backup source prepared", "path", prepared.Path)

	zipFile, err := c.compressBackup(prepared.Path)
	if err != nil {
		c.logStageError(stageName, "backup stage failed", err)
		c.report.EnsureFailed("备份失败")
		return err
	}
	defer func(path string) {
		err = os.Remove(path)
		if err != nil {
			c.logger.Error("zip cleanup failed", "file", path, "error", err)
			c.report.MarkError("清理zip文件失败")
		}
	}(zipFile)

	if conf.AfterCmd != "" {
		if err := c.runCommandStep("执行后置命令", conf.AfterCmd, "后置命令执行失败"); err != nil {
			c.logStageError(stageName, "backup stage failed", err)
			c.report.EnsureFailed("备份失败")
			return err
		}
	}

	if err := c.uploadBackup(zipFile); err != nil {
		c.logStageError(stageName, "backup stage failed", err)
		c.report.EnsureFailed("备份失败")
		return err
	}

	c.logStageFinish(stageName)
	return nil
}

func (c *TaskHolder) runCommandStep(stepName string, command string, errorMessage string) error {
	c.logStageStart(stepName)
	c.logger.Info("command executing", "stage", stepName, "command", command)

	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.logger.Error("command execution failed", "stage", stepName, "command", command, "error", err)
		c.report.MarkError(errorMessage)
		return err
	}

	c.logStageFinish(stepName)
	return nil
}

func (c *TaskHolder) compressBackup(path string) (string, error) {
	const stageName = "压缩文件"
	c.logStageStart(stageName)

	zipFile, err := utils.ZipPath(path, utils.GetFileName(c.ID), func(filePath string, processed, total int64, percentage float64) {
		c.logger.Info("compression progress", "stage", stageName, "file", filePath, "processed", notice.FormatBytes(processed), "total", notice.FormatBytes(total), "percentage", percentage)
	}, func(total int64) {
		c.report.SetCompressedSize(total)
		c.logger.Info("compression completed", "stage", stageName, "size", notice.FormatBytes(total), "bytes", total)
	})
	if err != nil {
		c.logger.Error("compression failed", "stage", stageName, "error", err)
		c.report.MarkError("压缩失败")
		return "", err
	}

	c.logStageFinish(stageName)
	return zipFile, nil
}

func (c *TaskHolder) uploadBackup(zipFile string) error {
	const stageName = "上传到OSS"
	objKey := filepath.Base(zipFile)
	ossClient := c.ossClient
	bucketName := ossClient.BucketName()

	c.logStageStart(stageName)
	c.logger.Info("upload started", "stage", stageName, "bucket", bucketName, "key", objKey)

	result, err := ossClient.Upload(objKey, zipFile)
	if err != nil {
		c.logger.Error("upload failed", "stage", stageName, "bucket", result.Bucket, "key", result.Key, "error", err)
		c.report.AddUploadFailure(result.Bucket, result.Key, err.Error())
		c.report.MarkError("上传失败")
		return err
	}

	c.logger.Info("upload succeeded", "stage", stageName, "bucket", result.Bucket, "key", result.Key, "mode", result.Mode)
	c.report.AddUploadSuccess(result.Bucket, result.Key)
	c.logStageFinish(stageName)
	return nil
}

func (c *TaskHolder) sendMessages() {
	c.noticeManager.NoticeReport(c.report.Snapshot())
}

func (c *TaskHolder) logStageStart(stageName string) {
	c.logger.Info("stage started", "stage", stageName)
}

func (c *TaskHolder) logStageFinish(stageName string) {
	c.logger.Info("stage completed", "stage", stageName)
}

func (c *TaskHolder) logStageError(stageName string, message string, err error) {
	c.logger.Error(message, "stage", stageName, "error", err)
}

func taskStatus(hasErrors bool) string {
	if hasErrors {
		return "failed"
	}
	return "success"
}
