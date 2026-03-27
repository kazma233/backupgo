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
	return c.logger.ExecuteStep("清理历史文件", func() error {
		deleted, err := c.ossClient.DeleteObjectsByPredicate(func(key string) bool {
			return utils.IsNeedDeleteFile(c.ID, key)
		})
		if err != nil {
			c.logger.LogError(err, "删除失败")
			return err
		}

		if len(deleted) == 0 {
			c.logger.LogInfo("无需删除文件")
			return nil
		}

		c.logger.LogInfo("成功删除：%v个文件", deleted)
		return nil
	})
}

func (c *TaskHolder) backup() error {
	conf := c.conf

	return c.logger.ExecuteStep("备份", func() error {
		if conf.BeforeCmd != "" {
			if err := c.runCommandStep("执行前置命令", conf.BeforeCmd, "前置命令执行失败"); err != nil {
				return err
			}
		}

		prepared, err := exporter.Prepare(c.ID, conf, c.logger)
		if err != nil {
			return err
		}
		defer prepared.Cleanup()

		c.logger.LogInfo("备份路径: %s", prepared.Path)

		zipFile, err := c.compressBackup(prepared.Path)
		if err != nil {
			return err
		}
		defer os.Remove(zipFile)

		if conf.AfterCmd != "" {
			if err := c.runCommandStep("执行后置命令", conf.AfterCmd, "后置命令执行失败"); err != nil {
				return err
			}
		}

		if err := c.uploadBackup(zipFile); err != nil {
			return err
		}

		return nil
	})
}

func (c *TaskHolder) runCommandStep(stepName string, command string, errorMessage string) error {
	return c.logger.ExecuteStep(stepName, func() error {
		c.logger.LogInfo("命令: %s", command)

		cmd := exec.Command("bash", "-c", command)
		if err := cmd.Run(); err != nil {
			c.logger.LogError(err, errorMessage)
			return err
		}

		return nil
	})
}

func (c *TaskHolder) compressBackup(path string) (string, error) {
	var zipFile string

	err := c.logger.ExecuteStep("压缩文件", func() error {
		var err error
		zipFile, err = utils.ZipPath(path, utils.GetFileName(c.ID), func(filePath string, processed, total int64, percentage float64) {
			c.logger.LogProgress(filePath, processed, total, percentage)
		}, func(total int64) {
			c.logger.LogCompressed(total)
		})
		if err != nil {
			c.logger.LogError(err, "压缩失败")
			return err
		}

		return nil
	})

	return zipFile, err
}

func (c *TaskHolder) uploadBackup(zipFile string) error {
	objKey := filepath.Base(zipFile)
	ossClient := c.ossClient

	return c.logger.ExecuteStep("上传到OSS", func() error {
		c.logger.LogInfo("文件: %s", objKey)

		err := ossClient.Upload(objKey, zipFile, func(status string) {
			c.logger.LogInfo("上传进度: %s", status)
		})

		if ossClient.HasError(err) {
			c.logger.LogError(err, "上传失败")
			return err
		}

		if ossClient.HasCoolDownError(err) {
			c.logger.LogInfo("上传因冷却期延迟: %s", objKey)
			return nil
		}

		c.logger.LogUpload("OSS", objKey)
		return nil
	})
}

func (c *TaskHolder) sendMessages() {
	c.noticeManager.NoticeEntries(c.ID, c.logger.GetStartTime(), c.logger.GetEntries())
}
