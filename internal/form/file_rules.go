package form

// File validation rules (separate from struct validation)
type FileRules struct {
	Required     bool
	MaxSize      int64    // bytes
	AllowedTypes []string // MIME types
	AllowedExts  []string // file extensions
	MinWidth     int      // for images
	MinHeight    int      // for images
	MaxWidth     int
	MaxHeight    int
}

var ListingFormFileRules = map[string]FileRules{
	"image": {
		Required:     true,
		MaxSize:      5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxWidth:     3000,
		MaxHeight:    3000,
	},
}

var VendorApplicationFormFileRules = map[string]FileRules{
	"logo": {
		Required:     true,
		MaxSize:      5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxWidth:     3000,
		MaxHeight:    3000,
	},
	"inventory": {
		Required:     false,
		MaxSize:      5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png", ".webp"},
		MaxWidth:     3000,
		MaxHeight:    3000,
	},
}
