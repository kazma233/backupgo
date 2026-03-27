package exporter

import (
	"path/filepath"

	"backupgo/config"
)

type mongoBackupSource struct {
	conf config.MongoBackupConfig
}

type mongoBackupAction struct {
	conf config.MongoBackupConfig
}

func (s mongoBackupSource) Prepare(taskID string, logger Logger) (*PreparedBackup, error) {
	return prepareCommandBackedBackup(taskID, logger, mongoBackupAction{conf: s.conf})
}

func (a mongoBackupAction) StepName() string {
	return "导出MongoDB"
}

func (a mongoBackupAction) ItemName() string {
	return "MongoDB 数据库"
}

func (a mongoBackupAction) Items() []string {
	return a.conf.Databases
}

func (a mongoBackupAction) TargetFile(rootDir string, item string) string {
	return filepath.Join(rootDir, mongoArchiveFileName(item, a.conf.Gzip))
}

func (a mongoBackupAction) Command(item string) commandSpec {
	return buildMongoDumpCommand(a.conf, item)
}

func buildMongoDumpCommand(conf config.MongoBackupConfig, database string) commandSpec {
	mongoArgs := []string{"--archive"}
	if conf.Gzip {
		mongoArgs = append(mongoArgs, "--gzip")
	}
	if conf.URI != "" {
		mongoArgs = appendStringOption(mongoArgs, "--uri", conf.URI)
	} else {
		mongoArgs = appendStringOption(mongoArgs, "--host", conf.Host)
		mongoArgs = appendIntOption(mongoArgs, "--port", conf.Port)
		mongoArgs = appendStringOption(mongoArgs, "--username", conf.Username)
		mongoArgs = appendStringOption(mongoArgs, "--password", conf.Password)
		mongoArgs = appendStringOption(mongoArgs, "--authenticationDatabase", conf.AuthDatabase)
	}
	mongoArgs = append(mongoArgs, conf.ExtraArgs...)
	mongoArgs = append(mongoArgs, "--db", database)

	if conf.GetMode() == config.ExecModeDocker {
		return dockerExecCommand(conf.Container, "mongodump", nil, mongoArgs)
	}

	return commandSpec{Name: "mongodump", Args: mongoArgs}
}

func mongoArchiveFileName(database string, gzip bool) string {
	name := sanitizeDumpFileName(database) + ".archive"
	if gzip {
		return name + ".gz"
	}
	return name
}
