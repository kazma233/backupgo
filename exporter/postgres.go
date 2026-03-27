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
	prepared, err := newPreparedData(s.taskID, s.logger)
	if err != nil {
		return nil, err
	}

	err = s.logger.ExecuteStep("导出Postgres", func() error {
		for _, db := range s.conf.Databases {
			targetFile := filepath.Join(prepared.Path, sanitizeDumpFileName(db)+".dump")
			s.logger.LogInfo("导出 Postgres 数据库 %s -> %s", db, targetFile)

			spec := buildPostgresDumpCommand(s.conf, db)
			if err := runCommandToFile(spec, targetFile); err != nil {
				s.logger.LogError(err, "Postgres 数据库 %s 导出失败", db)
				return err
			}
		}
		return nil
	})
	if err != nil {
		prepared.Cleanup()
		return nil, err
	}

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
