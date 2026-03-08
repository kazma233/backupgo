package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestFileNameProcessor_Generate(t *testing.T) {
	timestamp := time.Date(2024, 3, 8, 12, 0, 0, 0, time.UTC)

	out := defaultProcessor.Generate("test", timestamp)
	if out != "test_2024_03_08" {
		t.Fatalf("expected generated file name %q, got %q", "test_2024_03_08", out)
	}
}

func TestFileNameProcessor_Parse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *FNParserResult
		wantErr string
	}{
		{
			name:  "valid name with suffix",
			input: "test_cc_s_2024_03_08.zip",
			want: &FNParserResult{
				Prefix: "test_cc_s",
				Year:   2024,
				Month:  3,
				Day:    8,
			},
		},
		{
			name:    "invalid format",
			input:   "bad-name",
			wantErr: "invalid string format",
		},
		{
			name:    "invalid month",
			input:   "test_2024_13_08.zip",
			wantErr: "invalid month value",
		},
		{
			name:    "invalid day",
			input:   "test_2024_03_32.zip",
			wantErr: "invalid day value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultProcessor.Parse(tt.input)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected result %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestNeedDeleteFile(t *testing.T) {
	previousNow := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2024, 3, 10, 10, 0, 0, 0, time.UTC)
	}
	defer func() {
		nowFunc = previousNow
	}()

	oldName := defaultProcessor.Generate("test_cc_s", nowFunc().AddDate(0, 0, -9))
	if !IsNeedDeleteFile("test_cc_s", oldName) {
		t.Fatalf("expected %q to be deletable", oldName)
	}

	recentName := defaultProcessor.Generate("test_cc_s", nowFunc().AddDate(0, 0, -1))
	if IsNeedDeleteFile("test_cc_s", recentName) {
		t.Fatalf("expected %q to be kept", recentName)
	}

	if IsNeedDeleteFile("other_prefix", oldName) {
		t.Fatalf("expected prefix mismatch to return false")
	}

	if IsNeedDeleteFile("test_cc_s", "invalid-name") {
		t.Fatalf("expected invalid file name to return false")
	}
}

func TestGetFileName(t *testing.T) {
	previousNow := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2024, 3, 8, 12, 0, 0, 0, time.UTC)
	}
	defer func() {
		nowFunc = previousNow
	}()

	got := GetFileName("backup")
	if got != "backup_2024_03_08.zip" {
		t.Fatalf("expected file name %q, got %q", "backup_2024_03_08.zip", got)
	}
}
