package exporter

import (
	"fmt"

	"backupgo/config"
)

// Source 定义具体备份源的准备动作。
type Source interface {
	// PrepareData 根据任务配置生成可供压缩的备份产物。
	PrepareData() (*PreparedData, error)
}

// Prepare 根据任务配置选择备份源，并生成可供后续压缩的备份产物。
func Prepare(taskID string, conf config.BackupConfig, logger Logger) (*PreparedData, error) {
	source, err := New(taskID, conf, logger)
	if err != nil {
		return nil, err
	}

	return source.PrepareData()
}

// New 根据任务配置构造对应的备份源实现。
func New(taskID string, conf config.BackupConfig, logger Logger) (Source, error) {
	switch conf.GetType() {
	case config.BackupTypePath:
		return pathSource{taskID: taskID, logger: logger, path: conf.BackupPath}, nil
	case config.BackupTypePostgres:
		return postgresBackupSource{taskID: taskID, logger: logger, conf: *conf.Postgres}, nil
	case config.BackupTypeMongoDB:
		return mongoBackupSource{taskID: taskID, logger: logger, conf: *conf.MongoDB}, nil
	default:
		return nil, fmt.Errorf("unsupported backup type: %s", conf.GetType())
	}
}
