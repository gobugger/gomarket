package util

import (
	"errors"
	"fmt"
	_ "image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

var (
	ErrInvalidContentType = errors.New("invalid content type")
)

// Returns page slice and total number of pages or error
// Returns error if page number is invalid
func GetPage[T any](items []T, page, itemsPerPage int) ([]T, int, error) {
	if len(items) == 0 {
		return items, 1, nil
	}
	begin := (page - 1) * itemsPerPage
	end := min(begin+itemsPerPage, len(items))

	if begin >= end || begin >= len(items) || end > len(items) {
		return nil, 0, fmt.Errorf("invalid arguments")
	}

	numPages := 1 + len(items)/itemsPerPage
	if len(items)%itemsPerPage == 0 {
		numPages--
	}

	return items[begin:end], numPages, nil
}

type Closer interface {
	Close() error
}

func Close(c Closer) {
	if err := c.Close(); err != nil {
		slog.Error("failed to close Closer", slog.Any("error", err))
	}
}

func CheckContentType(file io.ReadSeeker, mimeTypePrefix string) error {
	// Read the first 512 bytes for MIME detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}

	contentType := http.DetectContentType(buf[:n])
	if !strings.HasPrefix(contentType, mimeTypePrefix) {
		return ErrInvalidContentType
	}

	// Reset file pointer for decoding
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	return nil
}
