package captcha

import (
	"github.com/gobugger/gomarket/pkg/rand"
	"image"
	"image/color"
	"math"
)

// nolint
func distance(a, b image.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2))
}

func drawCircle(img *image.RGBA, center image.Point, radius int, color color.RGBA, tickness int) {
	r := float64(radius)
	da := 1 / r / 10
	for a := 0.0; a < 6.3; a += da {
		cos := math.Cos(a)
		sin := math.Sin(a)
		rtmp := r
		for range tickness {
			x := int(rtmp*cos) + center.X
			y := int(rtmp*sin) + center.Y
			img.Set(x, y, color)
			rtmp--
		}
	}
}

func drawArch(img *image.RGBA, center image.Point, radius int, color color.RGBA, tickness int, startAngle float64, endAngle float64) {
	r := float64(radius)
	da := 1 / r / 10
	for a := startAngle; a < endAngle; a += da {
		cos := math.Cos(a)
		sin := math.Sin(a)
		rtmp := r
		for range tickness {
			x := int(rtmp*cos) + center.X
			y := int(rtmp*sin) + center.Y
			img.Set(x, y, color)
			rtmp--
		}
	}
}

func addNoise(img *image.RGBA, genColor func() color.RGBA) {
	for range 100 {
		x := rand.Intn(img.Bounds().Max.X)
		y := rand.Intn(img.Bounds().Max.Y)
		img.Set(x, y, genColor())
	}
}
