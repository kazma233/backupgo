package exporter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Logger 定义备份源执行过程中依赖的日志能力。
type Logger interface {
	// LogInfo 记录普通信息日志。
	LogInfo(format string, args ...interface{})
	// LogError 记录错误日志。
	LogError(err error, format string, args ...interface{})
}

type commandSpec struct {
	Name string
	Args []string
	Env  []string
}

var dumpFileNameCleaner = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

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
		return fmt.Errorf("%w: command failed without stderr output", err)
	}

	return nil
}

func runCommand(spec commandSpec) error {
	cmd := exec.Command(spec.Name, spec.Args...)
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}
	cmd.Stdout = io.Discard

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return fmt.Errorf("%w: %s", err, message)
		}
		return fmt.Errorf("%w: command failed without stderr output", err)
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
