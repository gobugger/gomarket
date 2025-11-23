package captcha

import (
	"encoding/gob"
	"image"
	"math"
)

type Answer struct {
	X int `schema:"x"`
	Y int `schema:"y"`
}

type Solution struct {
	Radius int
	X      int
	Y      int
}

type Captcha struct {
	Image    *image.RGBA
	Solution Solution
}

func New() Captcha {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{280, 140}})
	solution := createCircleCaptcha(img)

	return Captcha{
		Image:    img,
		Solution: solution,
	}
}

func (solution *Solution) Correct(answer Answer) bool {
	return math.Pow(float64(solution.X-answer.X), 2)+math.Pow(float64(solution.Y-answer.Y), 2) < math.Pow(float64(solution.Radius), 2)
}

func init() {
	gob.Register(Solution{})
}
