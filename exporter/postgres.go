package exporter

import (
	"log/slog"
	"path/filepath"

	"backupgo/config"
)

type postgresBackupSource struct {
	taskID string
	logger *slog.Logger
	conf   config.PostgresBackupConfig
}

func (s postgresBackupSource) PrepareData() (*PreparedData, error) {
	prepared, err := newPreparedData(s.taskID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("postgres export started")

	for _, db := range s.conf.Databases {
		targetFile := filepath.Join(prepared.Path, sanitizeDumpFileName(db)+".dump")
		s.logger.Info("postgres database export started", "database", db, "target_file", targetFile)

		spec := buildPostgresDumpCommand(s.conf, db)
		if err := runCommandToFile(spec, targetFile); err != nil {
			_ = prepared.Cleanup()
			s.logger.Error("postgres database export failed", "database", db, "error", err)
			return nil, err
		}
	}

	s.logger.Info("postgres export completed")
	return prepared, nil
}

func buildPostgresDumpCommand(conf config.PostgresBackupConfig, database string) commandSpec {
	pgArgs := []string{"--format=custom", "--no-password"}
	pgArgs = appendStringOption(pgArgs, "--host", conf.Host)
	pgArgs = appendIntOption(pgArgs, "--port", conf.Port)
	pgArgs = appendStringOption(pgArgs, "--username", conf.User)
	pgArgs = append(pgArgs, conf.ExtraArgs...)
	pgArgs = append(pgArgs, "--dbname", database)

	var env []string
	if conf.Password != "" {
		env = append(env, "PGPASSWORD="+conf.Password)
	}

	if conf.GetMode() == config.ExecModeDocker {
		return dockerExecCommand(conf.Container, "pg_dump", env, pgArgs)
	}

	spec := commandSpec{Name: "pg_dump", Args: pgArgs}
	spec.Env = env
	return spec
}
