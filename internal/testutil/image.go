package testutil

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func MockImage(t *testing.T, w, h int, format string) io.ReadSeeker {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})
	buf := bytes.NewBuffer([]byte{})
	var err error
	switch format {
	case "png":
		err = png.Encode(buf, img)
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, img, nil)
	default:
		t.Fatalf("invalid format %s\n", format)
	}
	require.NoError(t, err)
	return bytes.NewReader(buf.Bytes())
}
