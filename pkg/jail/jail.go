package jail

import (
	"github.com/disintegration/imaging"
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/pkg/rand"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/webp"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var face = basicfont.Face7x13

// Returns random indexes to the onion address (excluding the .onion)
func RandomIndexes() []int {
	noOnion, _ := strings.CutSuffix(config.OnionAddr, ".onion")
	return makeIndexesEasy(noOnion)
}

func GetSolution(indexes []int) string {
	onion := []rune(config.OnionAddr)
	solution := make([]rune, len(indexes))

	for i, idx := range indexes {
		solution[i] = onion[idx]
	}

	return string(solution)
}

/*
func drawRectangle(img *image.RGBA, rect image.Rectangle) {
	red := color.RGBA{255, 0, 0, 255}
	for x := rect.Min.X; x <= rect.Max.X; x++ {
		img.Set(x, rect.Min.Y, red)
		img.Set(x, rect.Max.Y, red)
	}

	for y := rect.Min.Y; y <= rect.Max.Y; y++ {
		img.Set(rect.Min.X, y, red)
		img.Set(rect.Max.X, y, red)
	}
}
*/

func GetImage(indexes []int) *image.NRGBA {
	onion := []rune(config.OnionAddr)
	for _, idx := range indexes {
		onion[idx] = ' '
	}

	text := string(onion)

	imgW, imgH := 280, 140
	onionW := face.Width * (len(config.OnionAddr)/2 + 5)
	onionH := face.Height
	x := rand.Intn(imgW-onionW-20) + 10
	y := rand.Intn(imgH-2*onionH-10) + 10
	rect := image.Rectangle{Min: image.Point{x - 5, y - 5}, Max: image.Point{x + onionW, y + 2*onionH + 8}}
	colors := []color.RGBA{
		{255, 0, 0, 255},    // Bright Red
		{255, 255, 0, 255},  // Bright Yellow
		{0, 255, 0, 255},    // Bright Green
		{255, 0, 255, 255},  // Bright Magenta
		{255, 165, 0, 255},  // Dark Orange
		{0, 0, 0, 255},      // Black
		{0, 255, 255, 255},  // Cyan
		{128, 0, 128, 255},  // Dark Purple
		{255, 20, 147, 255}, // Bright Pink
	}

	grabColor := func() color.RGBA {
		if len(colors) == 0 {
			return color.RGBA{0, 0, 0, 255}
		}
		i := rand.Intn(len(colors))
		c := colors[i]
		colors = slices.Delete(colors, i, i+1)
		return c
	}

	img := baseImage(image.Point{imgW, imgH}, rect, grabColor())

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + face.Height)},
	}

	dRed := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 0, 0, 255}),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + face.Height)},
	}

	missingOnly := ""
	for _, l := range text {
		if l == ' ' {
			missingOnly += "*"
		} else {
			missingOnly += " "
		}
	}

	n := len(text) / 2
	d.DrawString(text[0:n])
	d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + 2*face.Height)}
	d.DrawString(text[n:])

	n = len(missingOnly) / 2
	dRed.DrawString(missingOnly[0:n])
	dRed.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + 2*face.Height)}
	dRed.DrawString(missingOnly[n:])

	return img
}

// Retuns total of 4 indexes, two from beginning and two from end
func makeIndexesEasy(text string) (result []int) {
	result = append(result, rand.Intn(8))
	result = append(result, rand.Intn(8))

	result = append(result, len(text)-1-rand.Intn(8))
	result = append(result, len(text)-1-rand.Intn(8))

	m := map[int]bool{}
	for _, i := range result {
		m[i] = true
	}

	result = []int{}

	for k := range m {
		result = append(result, k)
	}

	slices.Sort(result)

	return
}

func baseImage(imageSize image.Point, cleanArea image.Rectangle, c color.RGBA) *image.NRGBA {
	word := config.SiteName

	var img *image.NRGBA
	if backgrounds != nil {
		img = backgrounds[rand.Intn(len(backgrounds))]
	} else {
		img = image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{imageSize.X, imageSize.Y}})
	}

	dst := image.NewNRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), img, image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: face,
	}

	i := 0
	n := 10
	for i < n {
		w := face.Width * (len(word) + 1)
		h := face.Height

		x := rand.Intn(imageSize.X - w)
		y := rand.Intn(imageSize.Y - h)
		rect := image.Rectangle{Min: image.Point{x, y}, Max: image.Point{x + w, y + h}}

		if rect.Overlaps(cleanArea) {
			continue
		}

		d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y + face.Height)}
		d.DrawString(word)
		i++
	}
	return dst
}

func loadBackgrounds(pattern string, w, h int) ([]*image.NRGBA, error) {
	names, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	backgrounds := []*image.NRGBA{}

	for _, name := range names {
		f, err := os.Open(name) // nolint
		if err != nil {
			return nil, err
		}
		defer util.Close(f)

		img, _, err := image.Decode(f)
		if err != nil {
			return nil, err
		}

		bg := imaging.Resize(img, w, h, imaging.Lanczos)

		backgrounds = append(backgrounds, bg)
	}

	return backgrounds, nil
}
