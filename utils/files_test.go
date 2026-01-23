package utils

import (
	"log"
	"os"
	"testing"
)

func Test_zipPath(t *testing.T) {
	target := "D:/test.zip"
	path, err := ZipPath(`F:\zip_demo`, target, func(filePath string, processed, total int64, percentage float64) {
		log.Printf("zip %s: %d/%d (%.2f%%)", filePath, processed, total, percentage)
	}, func(total int64) {
		log.Printf("zip done, total: %d", total)
	})
	if err != nil {
		panic(err)
	}

	log.Printf("path %s", path)
	os.Remove(target)
}
