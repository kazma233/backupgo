package task

import (
	"backupgo/config"
	"backupgo/exporter"
	"backupgo/notice"
	"backupgo/oss"
	"backupgo/state"
	"backupgo/utils"
	"os"
	"os/exec"
	"path/filepath"
)

type TaskHolder struct {
	ID            string
	conf          config.BackupConfig
	ossClient     *oss.OssClient
	noticeManager *notice.NoticeManager
	logger        *notice.TaskLogger
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
	}
	holder.initLogger()
	return holder
}

func (c *TaskHolder) initLogger() {
	c.logger = notice.NewTaskLogger(c.ID)
}

func (c *TaskHolder) BackupTask() {
	c.logger.StartNewTask()

	if err := c.backup(); err != nil {
		state.GetState().SetTaskRun(c.ID, "failed")
		c.sendMessages()
		return
	}

	state.GetState().SetTaskRun(c.ID, "success")

	if err := c.cleanHistory(); err != nil {
		state.GetState().SetTaskRun(c.ID, "failed")
	}

	c.sendMessages()
}

func (c *TaskHolder) cleanHistory() error {
	const stageName = "清理历史文件"
	c.logger.StartStage(stageName)

	deleted, err := c.ossClient.DeleteObjectsByPredicate(func(key string) bool {
		return utils.IsNeedDeleteFile(c.ID, key)
	})
	if err != nil {
		c.logger.LogError(err, "删除失败")
		c.logger.FailStage(stageName, err)
		return err
	}

	if len(deleted) == 0 {
		c.logger.LogInfo("无需删除文件")
		c.logger.FinishStage(stageName)
		return nil
	}

	c.logger.LogInfo("成功删除：%v个文件", deleted)
	c.logger.FinishStage(stageName)
	return nil
}

func (c *TaskHolder) backup() error {
	const stageName = "备份"
	conf := c.conf

	c.logger.StartStage(stageName)

	if conf.BeforeCmd != "" {
		if err := c.runCommandStep("执行前置命令", conf.BeforeCmd, "前置命令执行失败"); err != nil {
			c.logger.FailStage(stageName, err)
			return err
		}
	}

	prepared, err := exporter.Prepare(c.ID, conf, c.logger)
	if err != nil {
		c.logger.FailStage(stageName, err)
		return err
	}
	defer func() {
		const cleanupStageName = "清理临时文件"

		c.logger.StartStage(cleanupStageName)
		if cleanupErr := prepared.Cleanup(); cleanupErr != nil {
			c.logger.LogError(cleanupErr, "清理临时文件失败")
			c.logger.FailStage(cleanupStageName, cleanupErr)
		} else {
			c.logger.FinishStage(cleanupStageName)
		}

	}()

	c.logger.LogInfo("备份路径: %s", prepared.Path)

	zipFile, err := c.compressBackup(prepared.Path)
	if err != nil {
		c.logger.FailStage(stageName, err)
		return err
	}
	defer func(path string) {
		err = os.Remove(path)
		if err != nil {
			c.logger.LogError(err, "清理zip文件失败")
		}
	}(zipFile)

	if conf.AfterCmd != "" {
		if err := c.runCommandStep("执行后置命令", conf.AfterCmd, "后置命令执行失败"); err != nil {
			c.logger.FailStage(stageName, err)
			return err
		}
	}

	if err := c.uploadBackup(zipFile); err != nil {
		c.logger.FailStage(stageName, err)
		return err
	}

	c.logger.FinishStage(stageName)
	return nil
}

func (c *TaskHolder) runCommandStep(stepName string, command string, errorMessage string) error {
	c.logger.StartStage(stepName)
	c.logger.LogInfo("命令: %s", command)

	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.logger.LogError(err, errorMessage)
		c.logger.FailStage(stepName, err)
		return err
	}

	c.logger.FinishStage(stepName)
	return nil
}

func (c *TaskHolder) compressBackup(path string) (string, error) {
	const stageName = "压缩文件"
	c.logger.StartStage(stageName)

	zipFile, err := utils.ZipPath(path, utils.GetFileName(c.ID), func(filePath string, processed, total int64, percentage float64) {
		c.logger.LogProgress(filePath, processed, total, percentage)
	}, func(total int64) {
		c.logger.LogCompressed(total)
	})
	if err != nil {
		c.logger.LogError(err, "压缩失败")
		c.logger.FailStage(stageName, err)
		return "", err
	}

	c.logger.FinishStage(stageName)
	return zipFile, nil
}

func (c *TaskHolder) uploadBackup(zipFile string) error {
	const stageName = "上传到OSS"
	objKey := filepath.Base(zipFile)
	ossClient := c.ossClient

	c.logger.StartStage(stageName)
	c.logger.LogInfo("文件: %s", objKey)

	bt, err := ossClient.Upload(objKey, zipFile)
	if ossClient.HasError(err) {
		c.logger.LogInfo("使用 %s 上传失败，原因: %v", bt, err)
		c.logger.FailStage(stageName, err)
		return err
	}

	if ossClient.HasCoolDownError(err) {
		c.logger.LogInfo("上传失败，原因：上传因冷却期延迟: %s", objKey)
		c.logger.FinishStage(stageName)
		return nil
	}

	c.logger.LogUpload(ossClient.BucketName(), objKey)
	c.logger.FinishStage(stageName)
	return nil
}

func (c *TaskHolder) sendMessages() {
	c.noticeManager.NoticeSummary(c.logger.Summary())
}
