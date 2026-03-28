package exporter

import (
	"os"
	"path/filepath"

	"backupgo/config"
)

type dockerVolumeSource struct {
	taskID string
	logger Logger
	conf   config.DockerVolumeBackupConfig
}

func (s dockerVolumeSource) PrepareData() (*PreparedData, error) {
	prepared, err := newPreparedData(s.taskID)
	if err != nil {
		return nil, err
	}

	s.logger.LogInfo("开始导出 Docker volume")
	s.logger.LogInfo("检查 Docker volume: %s", s.conf.Volume)
	if err := runCommand(buildDockerVolumeInspectCommand(s.conf.Volume)); err != nil {
		_ = prepared.Cleanup()
		s.logger.LogError(err, "Docker volume %s 不存在或无法访问", s.conf.Volume)
		return nil, err
	}

	targetFile := filepath.Join(prepared.Path, dockerVolumeArchiveFileName(s.conf.Volume))
	s.logger.LogInfo("导出 Docker volume %s -> %s", s.conf.Volume, targetFile)
	s.logger.LogInfo("使用 helper 镜像: %s", s.conf.GetImage())

	if err := runCommand(buildDockerVolumeBackupCommand(s.conf, prepared.Path)); err != nil {
		_ = os.Remove(targetFile)
		_ = prepared.Cleanup()
		s.logger.LogError(err, "Docker volume %s 导出失败", s.conf.Volume)
		return nil, err
	}

	s.logger.LogInfo("Docker volume 导出完成")
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
