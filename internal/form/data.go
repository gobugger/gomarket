package form

import (
	"mime/multipart"
)

type FileData struct {
	Header   *multipart.FileHeader
	File     multipart.File
	Filename string
	Size     int64
	MimeType string
}

// FieldErrors maps field names to error messages
type FieldErrors map[string]string

// FormData wraps your form struct with errors and submitted values
type FormData[T any] struct {
	Data   T
	Files  map[string]*FileData
	Values map[string]string // For repopulating fields on error
	Errors FieldErrors
}

func NewFormData[T any]() *FormData[T] {
	return &FormData[T]{
		Errors: make(FieldErrors),
		Values: make(map[string]string),
		Files:  make(map[string]*FileData),
	}
}

func (f *FormData[T]) Close() {
	for _, fileData := range f.Files {
		if fileData.File != nil {
			fileData.File.Close() //nolint
		}
	}
}

func (f *FormData[T]) HasErrors() bool {
	return len(f.Errors) > 0
}
