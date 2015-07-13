// Package rotwriter provides an io.Writer that is being truncated to zero
// after having reached a certain size. The previous content is copied to a
// history file.
package rotwriter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultSize is used for a new log writer if no other size is being
	// specified (10 MB)
	DefaultSize = 10 * 1024 * 1024
)

type rotateWriter struct {
	mutex    sync.Mutex
	filename string
	file     *os.File
	maxSize  int64
}

// New creates a new rotate writer based on the specified file name. The file
// being rotated whenever the maximum size is being reached. If no maximum size
// is indicated (<=0) a default size of 10 MB is used. The rotated files use
// the same file name as the main file with an additional timestamp inserted
// before the extension.
func New(filename string, maxSize int64) (io.Writer, error) {
	if maxSize <= 0 {
		maxSize = DefaultSize
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	rw := &rotateWriter{
		filename: filename,
		file:     file,
		maxSize:  maxSize,
	}

	return rw, nil
}

func (rw *rotateWriter) Write(p []byte) (n int, err error) {
	stat, err := rw.file.Stat()
	if err == nil && stat.Size() > rw.maxSize {
		rw.file.Close()

		ext := filepath.Ext(rw.file.Name())
		base := strings.TrimSuffix(rw.file.Name(), ext)
		name := fmt.Sprintf("%s-%s%s", base, time.Now().Format("20060102-150405"), ext)

		err = os.Rename(rw.file.Name(), name)
		if err != nil {
			return 0, err
		}

		rw.file, err = os.OpenFile(rw.filename, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return 0, err
		}
	}

	return rw.file.Write(p)
}
