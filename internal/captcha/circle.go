package captcha

import (
	"github.com/gobugger/gomarket/pkg/rand"
	"image"
	"image/color"
	"math"
)

const radius = 20

func createCircleCaptcha(img *image.RGBA) Solution {
	tickness := 1
	numCircles := 6

	randCenter := func() func() image.Point {
		prev := image.Point{}
		return func() image.Point {
			for {
				p := image.Point{rand.Intn(img.Bounds().Max.X-2*radius) + radius, rand.Intn(img.Bounds().Max.Y-2*radius) + radius}
				if math.Pow(float64(prev.X-p.X), 2)+math.Pow(float64(prev.Y-p.Y), 2) > float64(radius*radius) {
					prev = p
					return p
				}
			}
		}
	}()

	randColor := func() color.RGBA {
		colors := []color.RGBA{
			{0, 0, 0, 255},
			{255, 0, 0, 255},
			{0, 255, 0, 255},
			{0, 0, 255, 255},
			{255, 255, 0, 255},
			{255, 0, 255, 255},
			{0, 255, 255, 255},
		}
		return colors[rand.Intn(len(colors))]
	}

	addNoise(img, randColor)

	centers := []image.Point{}
	for range numCircles {
		centers = append(centers, randCenter())
	}

	for _, center := range centers[:len(centers)-1] {
		drawCircle(img, center, radius, randColor(), tickness)
	}

	p := centers[len(centers)-1]
	start := rand.Float64() * math.Pi * 2
	end := start + 1.7*math.Pi
	drawArch(img, p, radius, randColor(), tickness, start, end)

	return Solution{
		Radius: radius,
		X:      p.X,
		Y:      p.Y,
	}
}
