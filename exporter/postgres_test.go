package exporter

import (
	"backupgo/config"
	"reflect"
	"testing"
)

func TestBuildPostgresDumpCommandLocal(t *testing.T) {
	spec := buildPostgresDumpCommand(config.PostgresBackupConfig{
		Host:      "127.0.0.1",
		Port:      5432,
		User:      "postgres",
		Password:  "secret",
		Databases: []string{"app"},
		ExtraArgs: []string{"--no-owner"},
	}, "app")

	if spec.Name != "pg_dump" {
		t.Fatalf("unexpected command name: %s", spec.Name)
	}

	wantArgs := []string{
		"--format=custom",
		"--no-password",
		"--host", "127.0.0.1",
		"--port", "5432",
		"--username", "postgres",
		"--no-owner",
		"--dbname", "app",
	}
	if !reflect.DeepEqual(spec.Args, wantArgs) {
		t.Fatalf("unexpected args: %#v", spec.Args)
	}

	wantEnv := []string{"PGPASSWORD=secret"}
	if !reflect.DeepEqual(spec.Env, wantEnv) {
		t.Fatalf("unexpected env: %#v", spec.Env)
	}
}

func TestBuildPostgresDumpCommandDocker(t *testing.T) {
	spec := buildPostgresDumpCommand(config.PostgresBackupConfig{
		Mode:      config.ExecModeDocker,
		Container: "postgres",
		User:      "postgres",
		Password:  "secret",
		Databases: []string{"app"},
	}, "app")

	if spec.Name != "docker" {
		t.Fatalf("unexpected command name: %s", spec.Name)
	}

	wantPrefix := []string{"exec", "-i", "-e", "PGPASSWORD=secret", "postgres", "pg_dump"}
	if !reflect.DeepEqual(spec.Args[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("unexpected docker args prefix: %#v", spec.Args)
	}
}
