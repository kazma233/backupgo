package exporter

import (
	"path/filepath"

	"backupgo/config"
)

type postgresBackupSource struct {
	taskID string
	logger Logger
	conf   config.PostgresBackupConfig
}

func (s postgresBackupSource) PrepareData() (*PreparedData, error) {
	prepared, err := newPreparedData(s.taskID)
	if err != nil {
		return nil, err
	}

	s.logger.LogInfo("开始导出 Postgres")

	for _, db := range s.conf.Databases {
		targetFile := filepath.Join(prepared.Path, sanitizeDumpFileName(db)+".dump")
		s.logger.LogInfo("导出 Postgres 数据库 %s -> %s", db, targetFile)

		spec := buildPostgresDumpCommand(s.conf, db)
		if err := runCommandToFile(spec, targetFile); err != nil {
			_ = prepared.Cleanup()
			s.logger.LogError(err, "Postgres 数据库 %s 导出失败", db)
			return nil, err
		}
	}

	s.logger.LogInfo("Postgres 导出完成")
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
