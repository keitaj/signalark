package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// recorder writes raw UBX bytes to timestamped files with optional rotation.
type recorder struct {
	dir      string
	rotate   time.Duration
	file     *os.File
	openedAt time.Time
}

func newRecorder(dir string, rotate time.Duration) (*recorder, error) {
	r := &recorder{dir: dir, rotate: rotate}
	if err := r.openNew(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *recorder) Write(p []byte) (int, error) {
	if r.rotate > 0 && time.Since(r.openedAt) >= r.rotate {
		if err := r.rotateFile(); err != nil {
			return 0, err
		}
	}
	return r.file.Write(p)
}

func (r *recorder) Close() error {
	if r.file != nil {
		r.file.Sync()
		return r.file.Close()
	}
	return nil
}

func (r *recorder) Path() string {
	if r.file != nil {
		return r.file.Name()
	}
	return ""
}

func (r *recorder) openNew() error {
	name := fmt.Sprintf("gnss_%s.ubx", time.Now().Format("20060102_150405"))
	path := filepath.Join(r.dir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	r.file = f
	r.openedAt = time.Now()
	fmt.Fprintf(os.Stderr, "  Recording to %s\n", path)
	return nil
}

func (r *recorder) rotateFile() error {
	if r.file != nil {
		r.file.Sync()
		r.file.Close()
	}
	return r.openNew()
}
