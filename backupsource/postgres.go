package backupsource

import (
	"path/filepath"

	"backupgo/config"
)

type postgresBackupSource struct {
	conf config.PostgresBackupConfig
}

type postgresBackupAction struct {
	conf config.PostgresBackupConfig
}

func (s postgresBackupSource) Prepare(taskID string, logger Logger) (*PreparedBackup, error) {
	return prepareCommandBackedBackup(taskID, logger, postgresBackupAction{conf: s.conf})
}

func (a postgresBackupAction) StepName() string {
	return "导出Postgres"
}

func (a postgresBackupAction) ItemName() string {
	return "Postgres 数据库"
}

func (a postgresBackupAction) Items() []string {
	return a.conf.Databases
}

func (a postgresBackupAction) TargetFile(rootDir string, item string) string {
	return filepath.Join(rootDir, sanitizeDumpFileName(item)+".dump")
}

func (a postgresBackupAction) Command(item string) commandSpec {
	return buildPostgresDumpCommand(a.conf, item)
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
