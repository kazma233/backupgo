package exporter

import (
	"fmt"
	"os"
	"path/filepath"
)

// PreparedData 表示已经准备完成、可用于后续压缩和上传的本地备份产物。
type PreparedData struct {
	Path    string
	cleanup func() error
	logger  Logger
}

// Cleanup 清理 Prepare 阶段生成的临时备份产物。
func (p *PreparedData) Cleanup() {
	if p == nil || p.cleanup == nil {
		return
	}

	if err := p.logger.ExecuteStep("清理临时文件", p.cleanup); err != nil {
		p.logger.LogError(err, "清理临时文件失败")
	}
}

func newPreparedData(taskID string, logger Logger) (*PreparedData, error) {
	sanitizedTaskID := sanitizeDumpFileName(taskID)

	rootDir, err := os.MkdirTemp("", "backupgo-"+sanitizedTaskID+"-")
	if err != nil {
		return nil, fmt.Errorf("create temp backup root failed: %w", err)
	}

	sourceDir := filepath.Join(rootDir, sanitizedTaskID)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		_ = os.RemoveAll(rootDir)
		return nil, fmt.Errorf("create temp backup dir failed: %w", err)
	}

	return &PreparedData{
		Path:   sourceDir,
		logger: logger,
		cleanup: func() error {
			return os.RemoveAll(rootDir)
		},
	}, nil
}
