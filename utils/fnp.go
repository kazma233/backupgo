package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	prefixMatchIndex = 1
	yearMatchIndex   = 2
	monthMatchIndex  = 3
	dayMatchIndex    = 4

	retentionDays = 7
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

var (
	defaultProcessor = &FileNameProcessor{
		rg:     regexp.MustCompile(`^([a-zA-Z0-9_]+)_(\d{4})_(\d{2})_(\d{2})`),
		format: `%s_%d_%02d_%02d`,
	}
	nowFunc = time.Now
)

// Generate 生成包含前缀和日期的字符串
func (sp *FileNameProcessor) Generate(prefix string, t time.Time) string {
	return fmt.Sprintf(sp.format, prefix, t.Year(), t.Month(), t.Day())
}

// Parse 解析包含前缀和日期的字符串，并返回充结构体
func (sp *FileNameProcessor) Parse(s string) (*FNParserResult, error) {
	matches := sp.rg.FindStringSubmatch(s)
	if matches == nil {
		return nil, errors.New("invalid string format")
	}

	prefix := matches[prefixMatchIndex]
	year, err := parseMatchInt(matches, yearMatchIndex)
	if err != nil {
		return nil, err
	}

	month, err := parseMatchIntInRange(matches, monthMatchIndex, 1, 12, "invalid month value")
	if err != nil {
		return nil, err
	}

	day, err := parseMatchIntInRange(matches, dayMatchIndex, 1, 31, "invalid day value")
	if err != nil {
		return nil, err
	}

	return &FNParserResult{
		Prefix: prefix,
		Year:   year,
		Month:  month,
		Day:    day,
	}, nil
}

func parseMatchInt(matches []string, index int) (int, error) {
	return strconv.Atoi(matches[index])
}

func parseMatchIntInRange(matches []string, index int, minValue int, maxValue int, errMessage string) (int, error) {
	value, err := parseMatchInt(matches, index)
	if err != nil {
		return 0, err
	}
	if value < minValue || value > maxValue {
		return 0, errors.New(errMessage)
	}

	return value, nil
}

func (r *FNParserResult) ToTime() time.Time {
	return time.Date(r.Year, time.Month(r.Month), r.Day, 0, 0, 0, 0, time.UTC)
}

func IsNeedDeleteFile(prefix, name string) bool {
	result, err := defaultProcessor.Parse(name)
	if err != nil {
		return false
	}
	if !strings.EqualFold(result.Prefix, prefix) {
		return false
	}

	fileDate := result.ToTime()
	beforeDate := nowFunc().AddDate(0, 0, -retentionDays)

	return fileDate.Before(beforeDate)
}

func GetFileName(prefix string) string {
	return defaultProcessor.Generate(prefix, nowFunc()) + ".zip"
}
