package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type ProgressCallback func(filePath string, processed int64, total int64, percentage float64)
type ProgressDoneCallback func(total int64)

// ProgressTracker 用于追踪进度
type ProgressTracker struct {
	processed    *int64
	total        int64
	callback     ProgressCallback
	currentFile  string
	done         chan bool
	doneCallback ProgressDoneCallback
}

func NewProgressTracker(total int64, callback ProgressCallback, doneCallback ProgressDoneCallback) *ProgressTracker {
	if callback == nil {
		panic("callback can not be nil")
	}

	var processed int64
	return &ProgressTracker{
		processed:    &processed,
		total:        total,
		callback:     callback,
		done:         make(chan bool),
		doneCallback: doneCallback,
	}
}

func (pt *ProgressTracker) Start() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				pt.callback(
					pt.currentFile,
					atomic.LoadInt64(pt.processed),
					pt.total,
					float64(atomic.LoadInt64(pt.processed))/float64(pt.total)*100,
				)
			case <-pt.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (pt *ProgressTracker) Stop() {
	close(pt.done)

	if pt.doneCallback != nil {
		pt.doneCallback(pt.total)
	}
}

func (pt *ProgressTracker) UpdateCurrentFile(path string) {
	pt.currentFile = path
}

func (pt *ProgressTracker) IncProcessed(size int) {
	atomic.AddInt64(pt.processed, int64(size))
}

func ZipPath(source string, target string, callback ProgressCallback, doneCallback ProgressDoneCallback) (string, error) {
	source = filepath.Clean(source)
	target = filepath.Clean(target)
	log.Printf("zip path: %s, target: %s", source, target)

	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat source path failed: %w", err)
	}

	if !info.IsDir() {
		return "", errors.New("source path is not a directory")
	}

	// 验证目标路径
	targetDir := filepath.Dir(target)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("create target directory failed: %w", err)
	}

	baseDir := filepath.Base(source)
	zipfile, err := os.Create(target)
	if err != nil {
		return "", fmt.Errorf("create zip file failed: %w", err)
	}
	defer zipfile.Close()

	// 计算总大小
	var totalSize int64
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("calculate total size failed: %w", err)
	}

	// 创建进度追踪器
	tracker := NewProgressTracker(totalSize, callback, doneCallback)
	tracker.Start()
	defer tracker.Stop()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk failed: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create file header failed: %w", err)
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("rel path failed: %w", err)
		}

		// Windows 路径分隔符转换为 ZIP 标准的 '/' 分隔符
		header.Name = filepath.ToSlash(filepath.Join(baseDir, relPath))
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create header failed: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		defer file.Close()

		tracker.UpdateCurrentFile(path)

		buf := make([]byte, 32*1024) // buffer
		for {
			nr, er := file.Read(buf)
			if nr > 0 {
				nw, ew := writer.Write(buf[:nr])
				if nw > 0 {
					tracker.IncProcessed(nw)
				}
				if ew != nil {
					return fmt.Errorf("write file failed: %w", ew)
				}
				if nw != nr {
					return fmt.Errorf("short write: wrote %d of %d bytes", nw, nr)
				}
			}
			if er == io.EOF {
				break
			}
			if er != nil {
				return fmt.Errorf("read file failed: %w", er)
			}
		}

		return nil
	})

	if err != nil {
		e := fmt.Errorf("zip failed: %w", err)
		return "", e
	}

	return target, nil
}
