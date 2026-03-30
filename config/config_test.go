package config

import "testing"

const testOSSConfig = `
oss:
  bucket_name: 'bucket'
  region: 'cn-hangzhou'
  access_key: 'access-key'
  access_key_secret: 'access-key-secret'
`

func withTestOSSConfig(config string) []byte {
	return []byte(testOSSConfig + config)
}

func TestParseConfigWithNoticeConfig(t *testing.T) {
	configBlob := withTestOSSConfig(`
notice:
  mail:
    smtp: 'smtp.example.com'
    port: 465
    user: 'user'
    password: 'password'
    to:
      - 'notice@example.com'
  telegram:
    bot_token: '123456:ABCDEF'
    chat_id: '123456789'
backup:
  - id: 'app'
    backup_path: './export'
`)

	cfg, err := ParseConfig(configBlob)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	if cfg.Notice == nil {
		t.Fatal("expected notice config to be present")
	}
	if cfg.Notice.Mail == nil {
		t.Fatal("expected mail config to be present")
	}
	if got := len(cfg.Notice.Mail.To); got != 1 {
		t.Fatalf("expected 1 mail recipient, got %d", got)
	}
	if cfg.Notice.Telegram == nil {
		t.Fatal("expected telegram config to be present")
	}
	if cfg.Notice.Telegram.BotToken != "123456:ABCDEF" {
		t.Fatalf("unexpected telegram bot token: %s", cfg.Notice.Telegram.BotToken)
	}
	if cfg.Notice.Telegram.ChatID != "123456789" {
		t.Fatalf("unexpected telegram chat id: %s", cfg.Notice.Telegram.ChatID)
	}
}

func TestParseConfigWithPostgresSource(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'pg'
    type: 'postgres'
    postgres:
      mode: 'docker'
      container: 'postgres'
      user: 'postgres'
      password: 'password'
      databases:
        - 'app'
`)

	cfg, err := ParseConfig(configBlob)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	task, ok := cfg.FindBackupByID("pg")
	if !ok {
		t.Fatal("expected pg task to be present")
	}
	if task.GetType() != BackupTypePostgres {
		t.Fatalf("unexpected backup type: %s", task.GetType())
	}
	if task.Postgres == nil {
		t.Fatal("expected postgres config to be present")
	}
	if got := task.Postgres.GetMode(); got != ExecModeDocker {
		t.Fatalf("unexpected postgres mode: %s", got)
	}
}

func TestParseConfigWithMongoSource(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'mongo'
    type: 'mongodb'
    mongodb:
      mode: 'local'
      uri: 'mongodb://root:pass@127.0.0.1:27017/?authSource=admin'
      gzip: true
      databases:
        - 'app'
`)

	cfg, err := ParseConfig(configBlob)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	task, ok := cfg.FindBackupByID("mongo")
	if !ok {
		t.Fatal("expected mongo task to be present")
	}
	if task.GetType() != BackupTypeMongoDB {
		t.Fatalf("unexpected backup type: %s", task.GetType())
	}
	if task.MongoDB == nil {
		t.Fatal("expected mongodb config to be present")
	}
	if !task.MongoDB.Gzip {
		t.Fatal("expected mongodb gzip to be enabled")
	}
}

func TestParseConfigWithDockerVolumeSource(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'docker-volume'
    docker_volume:
      volume: 'app-data'
`)

	cfg, err := ParseConfig(configBlob)
	if err != nil {
		t.Fatalf("ParseConfig returned error: %v", err)
	}

	task, ok := cfg.FindBackupByID("docker-volume")
	if !ok {
		t.Fatal("expected docker-volume task to be present")
	}
	if task.GetType() != BackupTypeDockerVolume {
		t.Fatalf("unexpected backup type: %s", task.GetType())
	}
	if task.DockerVolume == nil {
		t.Fatal("expected docker volume config to be present")
	}
	if task.DockerVolume.Volume != "app-data" {
		t.Fatalf("unexpected docker volume name: %s", task.DockerVolume.Volume)
	}
	if got := task.DockerVolume.GetImage(); got != "busybox:latest" {
		t.Fatalf("unexpected default docker volume image: %s", got)
	}
}

func TestParseConfigRejectsMultipleSources(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'invalid'
    type: 'postgres'
    backup_path: './export'
    postgres:
      databases:
        - 'app'
`)

	if _, err := ParseConfig(configBlob); err == nil {
		t.Fatal("expected ParseConfig to fail for multiple sources")
	}
}

func TestParseConfigRejectsDockerVolumeWithoutVolume(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'docker-volume'
    type: 'docker_volume'
    docker_volume:
      image: 'alpine:3.20'
`)

	if _, err := ParseConfig(configBlob); err == nil {
		t.Fatal("expected ParseConfig to fail for docker_volume without volume")
	}
}

func TestParseConfigRejectsDuplicateIDs(t *testing.T) {
	configBlob := withTestOSSConfig(`
backup:
  - id: 'dup'
    backup_path: './export-a'
  - id: 'dup'
    backup_path: './export-b'
`)

	if _, err := ParseConfig(configBlob); err == nil {
		t.Fatal("expected ParseConfig to fail for duplicate ids")
	}
}

func TestParseConfigRejectsMissingOSSConfig(t *testing.T) {
	configBlob := []byte(`
backup:
  - id: 'app'
    backup_path: './export'
`)

	if _, err := ParseConfig(configBlob); err == nil {
		t.Fatal("expected ParseConfig to fail for missing oss config")
	}
}

func TestParseConfigRejectsOSSWithoutRegion(t *testing.T) {
	configBlob := []byte(`
oss:
  bucket_name: 'bucket'
  access_key: 'access-key'
  access_key_secret: 'access-key-secret'
backup:
  - id: 'app'
    backup_path: './export'
`)

	if _, err := ParseConfig(configBlob); err == nil {
		t.Fatal("expected ParseConfig to fail for missing oss region")
	}
}
