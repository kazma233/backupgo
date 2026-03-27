package exporter

type pathSource struct {
	taskID string
	logger Logger
	path   string
}

func (s pathSource) PrepareData() (*PreparedData, error) {
	s.logger.LogInfo("使用目录备份源: %s", s.path)
	return &PreparedData{Path: s.path, logger: s.logger}, nil
}
