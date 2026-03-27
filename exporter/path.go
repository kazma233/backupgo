package exporter

type pathSource struct {
	path string
}

func (s pathSource) Prepare(taskID string, logger Logger) (*PreparedBackup, error) {
	logger.LogInfo("使用目录备份源: %s", s.path)
	return &PreparedBackup{Path: s.path}, nil
}
