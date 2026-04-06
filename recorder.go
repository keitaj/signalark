package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// recorder writes raw UBX bytes to timestamped files with optional rotation.
type recorder struct {
	dir        string
	rotate     time.Duration
	file       *os.File
	openedAt   time.Time
	written    int64 // total bytes written across all files
	writeErrs  int   // total write error count
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
	n, err := r.file.Write(p)
	r.written += int64(n)
	if err != nil {
		r.writeErrs++
		fmt.Fprintf(os.Stderr, "WARNING: raw write error (%d total): %v\n", r.writeErrs, err)
	}
	return n, err
}

func (r *recorder) Sync() error {
	if r.file != nil {
		return r.file.Sync()
	}
	return nil
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

// BytesWritten returns the total bytes written across all files.
func (r *recorder) BytesWritten() int64 { return r.written }

// WriteErrors returns the total number of write errors.
func (r *recorder) WriteErrors() int { return r.writeErrs }

// SizeString returns a human-readable string of the total bytes written.
func (r *recorder) SizeString() string {
	b := r.written
	switch {
	case b >= 1<<30:
		return fmt.Sprintf("%.2f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
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
