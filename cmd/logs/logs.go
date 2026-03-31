package logs

import (
	"backupgo/pkg/consts"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLogLines = 100
const tailReadBlockSize int64 = 4096

func LogsCommand() *cli.Command {
	return &cli.Command{
		Name:  "logs",
		Usage: "Print the last lines of the current scheduler log file",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "lines",
				Aliases: []string{"n"},
				Value:   defaultLogLines,
				Usage:   "Number of log lines to print",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runLogs(os.Stdout, cmd.Int("lines"))
		},
	}
}

func runLogs(output io.Writer, lineCount int) error {
	if lineCount < 0 {
		return fmt.Errorf("line count must be >= 0")
	}

	logFilePath, err := consts.LogFilePath()
	if err != nil {
		return err
	}

	return copyLogTail(output, logFilePath, lineCount)
}

func copyLogTail(output io.Writer, logFilePath string, lineCount int) error {
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("open log file %s failed: %w", logFilePath, err)
	}
	defer logFile.Close()

	if lineCount == 0 {
		return nil
	}

	startOffset, err := findTailStartOffset(logFile, lineCount)
	if err != nil {
		return fmt.Errorf("read log file %s failed: %w", logFilePath, err)
	}

	if _, err := logFile.Seek(startOffset, io.SeekStart); err != nil {
		return fmt.Errorf("seek log file %s failed: %w", logFilePath, err)
	}

	if _, err := io.Copy(output, logFile); err != nil {
		return fmt.Errorf("write log output failed: %w", err)
	}

	return nil
}

func findTailStartOffset(file *os.File, lineCount int) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}

	fileSize := info.Size()
	if fileSize == 0 || lineCount <= 0 {
		return 0, nil
	}

	scanEnd := fileSize
	lastByte := make([]byte, 1)
	if _, err := file.ReadAt(lastByte, fileSize-1); err != nil {
		return 0, err
	}
	if lastByte[0] == '\n' {
		scanEnd--
	}

	newlineCount := 0
	buffer := make([]byte, tailReadBlockSize)

	for offset := scanEnd; offset > 0; {
		readSize := tailReadBlockSize
		if offset < readSize {
			readSize = offset
		}

		start := offset - readSize
		chunk := buffer[:readSize]
		if _, err := file.ReadAt(chunk, start); err != nil {
			return 0, err
		}

		for i := len(chunk) - 1; i >= 0; i-- {
			if chunk[i] != '\n' {
				continue
			}

			newlineCount++
			if newlineCount == lineCount {
				return start + int64(i) + 1, nil
			}
		}

		offset = start
	}

	return 0, nil
}
