package exporter

import (
	"backupgo/config"
	"log/slog"
	"reflect"
	"testing"
)

func TestBuildDockerVolumeInspectCommand(t *testing.T) {
	spec := buildDockerVolumeInspectCommand("app-data")

	if spec.Name != "docker" {
		t.Fatalf("unexpected command name: %s", spec.Name)
	}

	wantArgs := []string{"volume", "inspect", "app-data"}
	if !reflect.DeepEqual(spec.Args, wantArgs) {
		t.Fatalf("unexpected args: %#v", spec.Args)
	}
}

func TestNewDockerVolumeSource(t *testing.T) {
	source, err := New("task", config.BackupConfig{
		Type: config.BackupTypeDockerVolume,
		DockerVolume: &config.DockerVolumeBackupConfig{
			Volume: "app-data",
		},
	}, slog.Default())
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, ok := source.(dockerVolumeSource); !ok {
		t.Fatalf("unexpected source type: %T", source)
	}
}

func TestBuildDockerVolumeBackupCommand(t *testing.T) {
	spec := buildDockerVolumeBackupCommand(config.DockerVolumeBackupConfig{
		Volume: "app-data",
		Image:  "alpine:3.20",
	}, "/tmp/backupgo-123/task")

	if spec.Name != "docker" {
		t.Fatalf("unexpected command name: %s", spec.Name)
	}

	wantArgs := []string{
		"run",
		"--rm",
		"--mount", "type=volume,src=app-data,dst=/source,readonly",
		"--mount", "type=bind,src=/tmp/backupgo-123/task,dst=/backup",
		"alpine:3.20",
		"tar",
		"-cf", "/backup/app-data.tar",
		"-C", "/source",
		".",
	}
	if !reflect.DeepEqual(spec.Args, wantArgs) {
		t.Fatalf("unexpected args: %#v", spec.Args)
	}
}

func TestBuildDockerVolumeBackupCommandUsesDefaultImage(t *testing.T) {
	spec := buildDockerVolumeBackupCommand(config.DockerVolumeBackupConfig{
		Volume: "app/data",
	}, "/tmp/out")

	if got := spec.Args[6]; got != "busybox:latest" {
		t.Fatalf("unexpected default image: %s", got)
	}
	if got := spec.Args[9]; got != "/backup/app_data.tar" {
		t.Fatalf("unexpected archive path: %s", got)
	}
}
