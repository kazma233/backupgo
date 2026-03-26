package backupsource

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"backupgo/config"
)

// PreparedBackup 表示已经准备完成、可用于后续压缩和上传的本地备份产物。
type PreparedBackup struct {
	Path    string
	Cleanup func() error
}

// Logger 定义备份源执行过程中依赖的日志能力。
type Logger interface {
	// ExecuteStep 包裹一个带步骤名的执行单元，并返回该步骤的执行结果。
	ExecuteStep(stepName string, fn func() error) error
	// LogInfo 记录普通信息日志。
	LogInfo(format string, args ...interface{})
	// LogError 记录错误日志。
	LogError(err error, format string, args ...interface{})
}

// Source 定义具体备份源的准备动作。
type Source interface {
	// Prepare 根据任务 ID 和日志对象生成可供压缩的备份产物。
	Prepare(taskID string, logger Logger) (*PreparedBackup, error)
}

type commandBackupAction interface {
	// StepName 返回该导出动作在日志中的步骤名。
	StepName() string
	// ItemName 返回当前导出对象的名称，用于拼接日志内容。
	ItemName() string
	// Items 返回本次需要导出的全部对象。
	Items() []string
	// TargetFile 返回单个导出对象对应的目标文件路径。
	TargetFile(rootDir string, item string) string
	// Command 返回执行单个导出对象所需的命令定义。
	Command(item string) commandSpec
}

type commandSpec struct {
	Name string
	Args []string
	Env  []string
}

var dumpFileNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// Prepare 根据任务配置选择备份源，并生成可供后续压缩的备份产物。
func Prepare(taskID string, conf config.BackupConfig, logger Logger) (*PreparedBackup, error) {
	source, err := New(conf)
	if err != nil {
		return nil, err
	}

	return source.Prepare(taskID, logger)
}

// New 根据任务配置构造对应的备份源实现。
func New(conf config.BackupConfig) (Source, error) {
	switch conf.GetType() {
	case config.BackupTypePath:
		return pathSource{path: conf.BackupPath}, nil
	case config.BackupTypePostgres:
		return postgresBackupSource{conf: *conf.Postgres}, nil
	case config.BackupTypeMongoDB:
		return mongoBackupSource{conf: *conf.MongoDB}, nil
	default:
		return nil, fmt.Errorf("unsupported backup type: %s", conf.GetType())
	}
}

// Cleanup 清理 Prepare 阶段生成的临时备份产物。
func Cleanup(logger Logger, prepared *PreparedBackup) {
	if prepared == nil || prepared.Cleanup == nil {
		return
	}

	if err := logger.ExecuteStep("清理临时文件", prepared.Cleanup); err != nil {
		logger.LogError(err, "清理临时文件失败")
	}
}

func prepareCommandBackedBackup(taskID string, logger Logger, action commandBackupAction) (*PreparedBackup, error) {
	prepared, err := newPreparedBackup(taskID)
	if err != nil {
		return nil, err
	}

	itemName := action.ItemName()
	items := action.Items()

	err = logger.ExecuteStep(action.StepName(), func() error {
		for _, item := range items {
			targetFile := action.TargetFile(prepared.Path, item)
			logger.LogInfo("导出 %s %s -> %s", itemName, item, targetFile)

			spec := action.Command(item)
			if err := runCommandToFile(spec, targetFile); err != nil {
				logger.LogError(err, "%s %s 导出失败", itemName, item)
				return err
			}
		}
		return nil
	})
	if err != nil {
		if cleanupErr := prepared.Cleanup(); cleanupErr != nil {
			logger.LogError(cleanupErr, "清理临时文件失败 (backup error: %v)", err)
		}
		return nil, err
	}

	return prepared, nil
}

func newPreparedBackup(taskID string) (*PreparedBackup, error) {
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

	return &PreparedBackup{
		Path: sourceDir,
		Cleanup: func() error {
			return os.RemoveAll(rootDir)
		},
	}, nil
}

func runCommandToFile(spec commandSpec, targetFile string) error {
	file, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("create target file failed: %w", err)
	}
	defer file.Close()

	cmd := exec.Command(spec.Name, spec.Args...)
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}
	cmd.Stdout = file

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = os.Remove(targetFile)
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return fmt.Errorf("%w: command exited without error message", err)
	}

	return nil
}

func sanitizeDumpFileName(value string) string {
	value = dumpFileNameCleaner.ReplaceAllString(value, "_")
	value = strings.Trim(value, "._-")
	if value == "" {
		return "dump"
	}
	return value
}

func appendStringOption(args []string, option string, value string) []string {
	if value == "" {
		return args
	}

	return append(args, option, value)
}

func appendIntOption(args []string, option string, value int) []string {
	if value <= 0 {
		return args
	}

	return append(args, option, strconv.Itoa(value))
}

func dockerExecCommand(container string, executable string, env []string, args []string) commandSpec {
	dockerArgs := []string{"exec", "-i"}
	for _, item := range env {
		dockerArgs = append(dockerArgs, "-e", item)
	}

	dockerArgs = append(dockerArgs, container, executable)
	dockerArgs = append(dockerArgs, args...)

	return commandSpec{Name: "docker", Args: dockerArgs}
}
