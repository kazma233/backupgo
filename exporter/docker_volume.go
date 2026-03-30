package exporter

import (
	"log/slog"
	"os"
	"path/filepath"

	"backupgo/config"
)

type dockerVolumeSource struct {
	taskID string
	logger *slog.Logger
	conf   config.DockerVolumeBackupConfig
}

func (s dockerVolumeSource) PrepareData() (*PreparedData, error) {
	prepared, err := newPreparedData(s.taskID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("docker volume export started", "volume", s.conf.Volume)
	s.logger.Info("docker volume inspect started", "volume", s.conf.Volume)
	if err := runCommand(buildDockerVolumeInspectCommand(s.conf.Volume)); err != nil {
		_ = prepared.Cleanup()
		s.logger.Error("docker volume inspect failed", "volume", s.conf.Volume, "error", err)
		return nil, err
	}

	targetFile := filepath.Join(prepared.Path, dockerVolumeArchiveFileName(s.conf.Volume))
	s.logger.Info("docker volume backup started", "volume", s.conf.Volume, "target_file", targetFile)
	s.logger.Info("docker volume helper image selected", "image", s.conf.GetImage())

	if err := runCommand(buildDockerVolumeBackupCommand(s.conf, prepared.Path)); err != nil {
		_ = os.Remove(targetFile)
		_ = prepared.Cleanup()
		s.logger.Error("docker volume backup failed", "volume", s.conf.Volume, "error", err)
		return nil, err
	}

	s.logger.Info("docker volume export completed", "volume", s.conf.Volume)
	return prepared, nil
}

func buildDockerVolumeInspectCommand(volume string) commandSpec {
	return commandSpec{
		Name: "docker",
		Args: []string{"volume", "inspect", volume},
	}
}

func buildDockerVolumeBackupCommand(conf config.DockerVolumeBackupConfig, outputDir string) commandSpec {
	archiveFile := dockerVolumeArchiveFileName(conf.Volume)

	return commandSpec{
		Name: "docker",
		Args: []string{
			"run",
			"--rm",
			"--mount", "type=volume,src=" + conf.Volume + ",dst=/source,readonly",
			"--mount", "type=bind,src=" + outputDir + ",dst=/backup",
			conf.GetImage(),
			"tar",
			"-cf", filepath.ToSlash(filepath.Join("/backup", archiveFile)),
			"-C", "/source",
			".",
		},
	}
}

func dockerVolumeArchiveFileName(volume string) string {
	return sanitizeDumpFileName(volume) + ".tar"
}
