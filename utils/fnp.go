package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FileNameProcessor struct {
	rg     *regexp.Regexp // match string
	format string
}

type FNParserResult struct {
	Prefix string
	Year   int
	Month  int
	Day    int
}

var defaultProcessor *FileNameProcessor

func init() {
	defaultProcessor = &FileNameProcessor{
		rg:     regexp.MustCompile(`^([a-zA-Z0-9_]+)_(\d{4})_(\d{2})_(\d{2})`),
		format: `%s_%d_%02d_%02d`,
	}
}

// GetDefaultProcessor returns the singleton instance
func GetDefaultProcessor() *FileNameProcessor {
	return defaultProcessor
}

// Generate 生成包含前缀和日期的字符串
func (sp *FileNameProcessor) Generate(prefix string, t time.Time) string {
	return fmt.Sprintf(sp.format, prefix, t.Year(), t.Month(), t.Day())
}

// Parse 解析包含前缀和日期的字符串，并返回充结构体
func (sp *FileNameProcessor) Parse(s string) (*FNParserResult, error) {
	// 正则表达式匹配前缀和日期，忽略后面的任何字符
	matches := sp.rg.FindStringSubmatch(s)

	if matches == nil {
		return nil, errors.New("invalid string format")
	}

	prefix := matches[1]
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	month, err := strconv.Atoi(matches[3])
	if err != nil || month < 1 || month > 12 {
		return nil, errors.New("invalid month value")
	}

	day, err := strconv.Atoi(matches[4])
	if err != nil || day < 1 || day > 31 {
		return nil, errors.New("invalid day value")
	}

	return &FNParserResult{
		Prefix: prefix,
		Year:   year,
		Month:  month,
		Day:    day,
	}, nil
}

func (r *FNParserResult) ToTime() time.Time {
	return time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)
}

func IsNeedDeleteFile(prefix, name string) bool {
	result, err := GetDefaultProcessor().Parse(name)
	if err != nil || !strings.EqualFold(result.Prefix, prefix) {
		return false
	}

	fileDate := result.ToTime()
	beforeDate := time.Now().AddDate(0, 0, -7)

	return fileDate.Before(beforeDate)
}

func GetFileName(prefix string) string {
	return GetDefaultProcessor().Generate(prefix, time.Now()) + ".zip"
}
