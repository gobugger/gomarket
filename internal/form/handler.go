package form

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gobugger/gomarket/internal/captcha"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gorilla/schema"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
)

var (
	_decoder   *schema.Decoder
	_validator *validator.Validate
)

type FormHandler[T any] struct {
	decoder   *schema.Decoder
	validator *validator.Validate
	fileRules map[string]FileRules
	maxMemory int64 // for ParseMultipartForm
}

func NewFormHandler[T any]() *FormHandler[T] {
	return &FormHandler[T]{
		decoder:   _decoder,
		validator: _validator,
		fileRules: make(map[string]FileRules),
		maxMemory: 32 << 20, // 32MB default
	}
}

func (h *FormHandler[T]) WithFileRules(rules map[string]FileRules) *FormHandler[T] {
	h.fileRules = rules
	return h
}

func (h *FormHandler[T]) WithMaxMemory(bytes int64) *FormHandler[T] {
	h.maxMemory = bytes
	return h
}

// Returns error only for unexpected errors not on validation errors
// Validation errors are written to FormData.Errors
func (h *FormHandler[T]) Parse(r *http.Request, cs *captcha.Solution) (*FormData[T], error) {
	// Determine if multipart
	contentType := r.Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

	formData := NewFormData[T]()

	closeFormData := true
	defer func() {
		if closeFormData {
			formData.Close()
		}
	}()

	if isMultipart {
		if err := r.ParseMultipartForm(h.maxMemory); err != nil {
			return nil, fmt.Errorf("failed to parse multipart form: %w", err)
		}

		// Process files
		if r.MultipartForm != nil {
			for fieldName, rules := range h.fileRules {
				h.processFileField(r, formData, fieldName, rules)
			}
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return nil, err
		}
	}

	// Store text field values for repopulation
	for key, values := range r.Form {
		if len(values) > 0 {
			formData.Values[key] = values[0]
		}
	}

	// Decode form data
	if err := h.decoder.Decode(&formData.Data, r.Form); err != nil {
		formData.Errors["_form"] = "Invalid form data"
		return formData, err
	}

	// Check captcha if used
	if config.CaptchaEnabled() {
		if c, ok := any(&formData.Data).(CaptchaValidable); ok {
			if cs == nil {
				return formData, fmt.Errorf("missing captcha solution")
			}
			if !cs.Correct(c.Answer()) {
				formData.Errors["captcha_answer"] = "invalid answer"
			}
		}
	}

	// Validate struct fields
	if err := h.validator.Struct(formData.Data); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				fieldName := camel2snake(e.Field())
				formData.Errors[fieldName] = formatError(e)
			}
		}
	}

	closeFormData = false

	return formData, nil
}

func (h *FormHandler[T]) processFileField(
	r *http.Request,
	formData *FormData[T],
	fieldName string,
	rules FileRules,
) {
	file, header, err := r.FormFile(fieldName)

	// Check if file is required
	if err == http.ErrMissingFile {
		if rules.Required {
			formData.Errors[fieldName] = fmt.Sprintf("%s is required", fieldName)
		}
		return
	}

	if err != nil {
		formData.Errors[fieldName] = "Failed to process file"
		return
	}

	closeFile := true
	defer func() {
		if closeFile {
			file.Close()
		}
	}()

	// Validate file size
	if rules.MaxSize > 0 && header.Size > rules.MaxSize {
		formData.Errors[fieldName] = fmt.Sprintf(
			"File size must be less than %s",
			humanizeBytes(rules.MaxSize),
		)
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if len(rules.AllowedExts) > 0 && !slices.Contains(rules.AllowedExts, ext) {
		formData.Errors[fieldName] = fmt.Sprintf(
			"Only %s files are allowed",
			strings.Join(rules.AllowedExts, ", "),
		)
		return
	}

	// Detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		formData.Errors[fieldName] = "Failed to read file"
		return
	}
	file.Seek(0, 0) // Reset file pointer

	mimeType := http.DetectContentType(buffer)

	// Validate MIME type
	if len(rules.AllowedTypes) > 0 &&
		!slices.ContainsFunc(rules.AllowedTypes, func(allowedType string) bool {
			return strings.HasPrefix(mimeType, allowedType)
		}) {
		formData.Errors[fieldName] = "Invalid file type"
		return
	}

	// Validate image dimensions if image
	if strings.HasPrefix(mimeType, "image/") &&
		(rules.MinWidth > 0 || rules.MinHeight > 0 || rules.MaxWidth > 0 || rules.MaxHeight > 0) {
		if err := h.validateImageDimensions(file, fieldName, rules, formData); err != nil {
			return
		}
		file.Seek(0, 0) // Reset after reading image
	}

	// We got to the end so don't close file
	closeFile = false

	// Store file data
	formData.Files[fieldName] = &FileData{
		Header:   header,
		File:     file,
		Filename: header.Filename,
		Size:     header.Size,
		MimeType: mimeType,
	}
}

func (h *FormHandler[T]) validateImageDimensions(
	file multipart.File,
	fieldName string,
	rules FileRules,
	formData *FormData[T],
) error {
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		formData.Errors[fieldName] = "Invalid image file"
		return err
	}

	if rules.MinWidth > 0 && img.Width < rules.MinWidth {
		formData.Errors[fieldName] = fmt.Sprintf(
			"Image width must be at least %dpx", rules.MinWidth,
		)
		return fmt.Errorf("width too small")
	}

	if rules.MinHeight > 0 && img.Height < rules.MinHeight {
		formData.Errors[fieldName] = fmt.Sprintf(
			"Image height must be at least %dpx", rules.MinHeight,
		)
		return fmt.Errorf("height too small")
	}

	if rules.MaxWidth > 0 && img.Width > rules.MaxWidth {
		formData.Errors[fieldName] = fmt.Sprintf(
			"Image width must be at most %dpx", rules.MaxWidth,
		)
		return fmt.Errorf("width too large")
	}

	if rules.MaxHeight > 0 && img.Height > rules.MaxHeight {
		formData.Errors[fieldName] = fmt.Sprintf(
			"Image height must be at most %dpx", rules.MaxHeight,
		)
		return fmt.Errorf("height too large")
	}

	return nil
}

func humanizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	_decoder = schema.NewDecoder()
	_decoder.IgnoreUnknownKeys(true)

	_validator = validator.New()
	_validator.RegisterValidation("location", location)
	_validator.RegisterValidation("xmr_address", xmrAddress)
	_validator.RegisterValidation("locale", locale)
	_validator.RegisterValidation("currency", currency)
}
