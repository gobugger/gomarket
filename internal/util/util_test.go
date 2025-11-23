package util

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/gobugger/gomarket/internal/testutil"
	"image"
	"image/color"
	"slices"
	"testing"
)

func TestGetPage(t *testing.T) {
	items := []int{}
	for i := range 13 {
		items = append(items, i)
	}

	res, n, err := GetPage(items, 1, 9)
	if err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatalf("n == %d\n", n)
	} else if !slices.Equal(res, items[0:9]) {
		t.Fatalf("res is invalid %v\n", res)
	}

	res, n, err = GetPage(items, 2, 9)
	if err != nil {
		t.Fatal(err)
	} else if n != 2 {
		t.Fatalf("n == %d\n", n)
	} else if !slices.Equal(res, items[9:]) {
		t.Fatalf("res is invalid %v\n", res)
	}
}

func TestSaveAndLoadImage(t *testing.T) {
	ctx := t.Context()
	mc, err := testutil.NewMinioContainer(ctx)
	defer mc.Terminate(ctx)

	src := testutil.MockImage(t, 500, 500, "jpeg")
	name := uuid.New().String() + ".jpeg"

	img, err := DecodeImage(src)
	require.NoError(t, err)

	img = TransformImage(img, 400, 400)

	err = SaveImage(ctx, mc.Client, "bucket", name, img)
	require.NoError(t, err)

	imgOut, err := LoadImage(ctx, mc.Client, "bucket", name)
	require.NoError(t, err)
	require.Equal(t, 400, imgOut.Bounds().Max.X)
	require.Equal(t, 400, imgOut.Bounds().Max.Y)
}

func TestTransformImage(t *testing.T) {
	tests := []struct {
		inW      int
		inH      int
		inFormat string
		outW     int
		outH     int
		fail     bool
	}{
		{
			1920, 1090, "jpeg",
			500, 500,
			false,
		},
		{
			1920, 1090, "jpeg",
			200, 200,
			false,
		},
		{
			1920, 1090, "png",
			1000, 1100,
			false,
		},
		{
			1920, 1090, "jpg",
			500, 0,
			false,
		},
		{
			1920, 1090, "jpeg",
			0, 0,
			false,
		},
		{
			200, 200, "png",
			500, 500,
			false,
		},
		{
			5001, 5001, "jpeg",
			500, 500,
			true,
		},
		{
			5001, 500, "png",
			500, 500,
			true,
		},
		{
			500, 5001, "jpeg",
			500, 500,
			true,
		},
		{
			500, 5001, "jpeg",
			500, 500,
			true,
		},
	}

	for _, test := range tests {
		jpgImage := testutil.MockImage(t, test.inW, test.inH, test.inFormat)

		img, err := DecodeImage(jpgImage)
		require.NoError(t, err)

		img = TransformImage(img, test.outW, test.outH)

		outW := img.Bounds().Max.X - img.Bounds().Min.X
		outH := img.Bounds().Max.Y - img.Bounds().Min.Y

		if test.outW != 0 {
			require.Equal(t, test.outW, outW)
		}
		if test.outH != 0 {
			require.Equal(t, test.outH, outH)
		}

		if test.outW == 0 || test.outH == 0 {
			require.InDelta(t, float64(test.inW)/float64(test.inH), float64(outW)/float64(outH), 0.01)
		}
	}
}

func BenchmarkBase64Encode(b *testing.B) {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{1920, 1080}})

	for i := range 1000 {
		img.SetRGBA(i, i, color.RGBA{1, 100, 255, 255})
	}

	buf := &bytes.Buffer{}
	for b.Loop() {
		Base64Encode(buf, img)
	}
}
