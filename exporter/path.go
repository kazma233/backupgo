package exporter

import "log/slog"

type pathSource struct {
	taskID string
	logger *slog.Logger
	path   string
}

func (s pathSource) PrepareData() (*PreparedData, error) {
	s.logger.Info("using path backup source", "path", s.path)
	return &PreparedData{Path: s.path}, nil
}
