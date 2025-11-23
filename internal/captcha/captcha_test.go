package captcha

import (
	"image"
	"image/color"
	"testing"

	"github.com/alexedwards/scs/v2"
)

func BenchmarkRefresh(b *testing.B) {
	cache := make([]item, cacheSize)
	for b.Loop() {
		refreshCache(cache)
	}
}

func BenchmarkTemplateField(b *testing.B) {
	ctx := b.Context()
	s := scs.New()
	ctx, _ = s.Load(ctx, "testsession")
	for b.Loop() {
		TemplateField(ctx, s)
	}
}

func BenchmarkDrawCircle(b *testing.B) {
	w, h := 100, 100
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{w, h}})
	center := image.Point{w / 2, h / 2}
	color := color.RGBA{255, 0, 0, 255}
	for b.Loop() {
		drawCircle(img, center, radius, color, 1)
	}
}
