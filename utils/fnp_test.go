package utils

import (
	"testing"
	"time"
)

func TestFileNameProcessor_Generate(t *testing.T) {
	out := GetDefaultProcessor().Generate("test", time.Now())
	t.Logf("result %v", out)
}

func TestNeedDeleteFile(t *testing.T) {
	out := GetDefaultProcessor().Generate("test_cc_s", time.Now().AddDate(0, 0, -9))
	t.Logf("Generate %v", out)
	res := IsNeedDeleteFile("test_cc_s", out)
	t.Logf("NeedDeleteFile %v", res)
}
