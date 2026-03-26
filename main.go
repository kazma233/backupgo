package main

import (
	"backupgo/backupsource"
	"backupgo/config"
	"backupgo/notice"
	"backupgo/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
)

type TaskHolder struct {
	ID            string
	conf          config.BackupConfig
	ossClient     *OssClient
	noticeManager *notice.NoticeManager
	logger        *notice.TaskLogger
}

func newTaskHolder(conf config.BackupConfig) *TaskHolder {
	if err := conf.Validate(); err != nil {
		panic(err)
	}

	holder := &TaskHolder{
		ID:            conf.GetID(),
		conf:          conf,
		ossClient:     CreateOSSClient(config.Config.OSS),
		noticeManager: notice.NewManagerFromConfig(config.Config),
	}
	holder.initLogger()
	return holder
}

func main() {
	config.InitConfig()

	secondParser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor,
	)
	c := cron.New(cron.WithParser(secondParser), cron.WithChain())

	for _, conf := range config.Config.BackupConf {
		dh := newTaskHolder(conf)

		backupTaskCron := conf.BackupTask
		if backupTaskCron == "" {
			backupTaskCron = "0 25 0 * * ?"
		}
		taskId, err := c.AddFunc(backupTaskCron, func() {
			dh.backupTask()
		})
		if err != nil {
			panic(err)
		}

		log.Printf("task %v add success", taskId)
	}

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		conf, ok := config.Config.FindBackupByID(id)
		if !ok {
			http.Error(resp, "backup task not found", http.StatusNotFound)
			return
		}

		dh := newTaskHolder(conf)
		log.Printf("backup task %v", dh)

		dh.backupTask()
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ids": config.Config.BackupIDs(),
		})
	})

	log.Println(http.ListenAndServe(":7000", nil))
}

func (c *TaskHolder) initLogger() {
	c.logger = notice.NewTaskLogger(c.ID)
}

func (c *TaskHolder) backupTask() {
	c.logger.StartNewTask()

	c.logger.ExecuteStep("BackupTask", func() error {
		var taskErr error

		if err := c.backupWithLogger(); err != nil {
			taskErr = err
		}

		if err := c.cleanHistoryWithLogger(); err != nil && taskErr == nil {
			taskErr = err
		}

		return taskErr
	})

	c.sendMessages()
}

func (c *TaskHolder) cleanHistory() {
	c.cleanHistoryWithLogger()
}

func (c *TaskHolder) cleanHistoryWithLogger() error {
	return c.logger.ExecuteStep("清理历史文件", func() error {
		ossClient := c.ossClient

		var objects []oss.ObjectProperties
		token := ""
		for {
			resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
			if err != nil {
				c.logger.LogError(err, "列出对象失败")
				return err
			}

			for _, object := range resp.Objects {
				need := utils.IsNeedDeleteFile(c.ID, object.Key)
				if need {
					objects = append(objects, object)
				}
			}
			if resp.IsTruncated {
				token = resp.NextContinuationToken
			} else {
				break
			}
		}

		if len(objects) <= 0 {
			c.logger.LogInfo("无需删除文件")
			return nil
		}

		var keys []string
		for _, k := range objects {
			keys = append(keys, k.Key)
		}

		c.logger.LogInfo("找到 %d 个文件需要删除", len(keys))
		deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
		if err != nil {
			c.logger.LogError(err, "删除失败")
			return err
		}

		c.logger.LogInfo("成功删除：%v", deleteObjects.DeletedObjects)
		return nil
	})
}

func (c *TaskHolder) backupWithLogger() error {
	conf := c.conf

	return c.logger.ExecuteStep("备份", func() error {
		if conf.BeforeCmd != "" {
			if err := c.runCommandStep("执行前置命令", conf.BeforeCmd, "前置命令执行失败"); err != nil {
				return err
			}
		}

		prepared, err := backupsource.Prepare(c.ID, conf, c.logger)
		if err != nil {
			return err
		}
		defer backupsource.Cleanup(c.logger, prepared)

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

// sendMessages 发送 TaskLogger 收集的所有消息
func (c *TaskHolder) sendMessages() {
	c.noticeManager.NoticeEntries(c.ID, c.logger.GetStartTime(), c.logger.GetEntries())
}
