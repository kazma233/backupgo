package logs

import (
	"backupgo/pkg/consts"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

const defaultLogLines = 100

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

	lines, err := tailLinesFromReader(logFile, lineCount)
	if err != nil {
		return fmt.Errorf("read log file %s failed: %w", logFilePath, err)
	}

	for _, line := range lines {
		if _, err := io.WriteString(output, line); err != nil {
			return fmt.Errorf("write log output failed: %w", err)
		}
	}

	return nil
}

func tailLinesFromReader(input io.Reader, lineCount int) ([]string, error) {
	if lineCount == 0 {
		return nil, nil
	}

	reader := bufio.NewReader(input)
	ring := make([]string, lineCount)
	total := 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if line != "" {
					ring[total%lineCount] = line
					total++
				}
				break
			}
			return nil, err
		}

		ring[total%lineCount] = line
		total++
	}

	if total == 0 {
		return nil, nil
	}

	count := total
	if count > lineCount {
		count = lineCount
	}

	start := 0
	if total > lineCount {
		start = total % lineCount
	}

	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lines = append(lines, ring[(start+i)%lineCount])
	}

	return lines, nil
}
