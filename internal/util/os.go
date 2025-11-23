package util

import (
	"io"
	"os"
)

func OpenFile(root, name string) (*os.File, error) {
	file, err := os.OpenInRoot(root, name)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func ReadFile(root, name string) ([]byte, error) {
	file, err := OpenFile(root, name)
	if err != nil {
		return nil, err
	}
	defer Close(file)

	return io.ReadAll(file)
}
