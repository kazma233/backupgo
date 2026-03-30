package exporter

import (
	"log/slog"
	"path/filepath"

	"backupgo/config"
)

type mongoBackupSource struct {
	taskID string
	logger *slog.Logger
	conf   config.MongoBackupConfig
}

func (s mongoBackupSource) PrepareData() (*PreparedData, error) {
	prepared, err := newPreparedData(s.taskID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("mongodb export started")

	for _, db := range s.conf.Databases {
		targetFile := filepath.Join(prepared.Path, mongoArchiveFileName(db, s.conf.Gzip))
		s.logger.Info("mongodb database export started", "database", db, "target_file", targetFile)

		spec := buildMongoDumpCommand(s.conf, db)
		if err := runCommandToFile(spec, targetFile); err != nil {
			_ = prepared.Cleanup()
			s.logger.Error("mongodb database export failed", "database", db, "error", err)
			return nil, err
		}
	}

	s.logger.Info("mongodb export completed")
	return prepared, nil
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
