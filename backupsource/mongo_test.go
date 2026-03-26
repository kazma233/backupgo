package backupsource

import (
	"backupgo/config"
	"reflect"
	"testing"
)

func TestBuildMongoDumpCommandDocker(t *testing.T) {
	spec := buildMongoDumpCommand(config.MongoBackupConfig{
		Mode:         config.ExecModeDocker,
		Container:    "mongo",
		Username:     "root",
		Password:     "secret",
		AuthDatabase: "admin",
		Gzip:         true,
		Databases:    []string{"app"},
	}, "app")

	if spec.Name != "docker" {
		t.Fatalf("unexpected command name: %s", spec.Name)
	}

	wantArgs := []string{
		"exec", "-i", "mongo", "mongodump",
		"--archive",
		"--gzip",
		"--username", "root",
		"--password", "secret",
		"--authenticationDatabase", "admin",
		"--db", "app",
	}
	if !reflect.DeepEqual(spec.Args, wantArgs) {
		t.Fatalf("unexpected args: %#v", spec.Args)
	}
}

func TestMongoArchiveFileName(t *testing.T) {
	if got := mongoArchiveFileName("app/logs", true); got != "app_logs.archive.gz" {
		t.Fatalf("unexpected archive name: %s", got)
	}
}
