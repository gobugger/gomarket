package form

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/gobugger/gomarket/internal/captcha"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestFormatField(t *testing.T) {
	require.Equal(t, "password", formatField("Password"))
	require.Equal(t, "password check", formatField("PasswordCheck"))
}

func TestCamel2Snake(t *testing.T) {
	require.Equal(t, "password", camel2snake("password"))
	require.Equal(t, "password", camel2snake("Password"))
	require.Equal(t, "passwor_d", camel2snake("PassworD"))
	require.Equal(t, "password_check", camel2snake("PasswordCheck"))
	require.Equal(t, "password_check", camel2snake("Password_Check"))
}

func TestForms(t *testing.T) {
	tests := []struct {
		name string
		form any
	}{
		{"register", any(&RegisterForm{})},
		{"jail", any(&JailForm{})},
	}

	for _, test := range tests {
		_, ok := any(test.form).(CaptchaValidable)
		require.True(t, ok, test.name)
	}
}

func TestRegisterForm(t *testing.T) {
	tests := []struct {
		username string
		password string
		check    string
		x        int
		y        int
		ferrors  FieldErrors
	}{
		{
			username: "mcafee",
			password: "secretpass123",
			check:    "secretpass123",
			x:        6,
			y:        9,
			ferrors:  FieldErrors{},
		},
		{
			username: "mcafee",
			password: "secretpass123",
			check:    "secretpass12",
			x:        6,
			y:        9,
			ferrors: FieldErrors{
				"PasswordCheck": "password check must match password",
			},
		},
		{
			username: "mcafee",
			password: "secretpass123",
			check:    "secretpass123",
			x:        200,
			y:        9,
			ferrors: FieldErrors{
				"CaptchaAnswer": "invalid answer",
			},
		},
		{
			username: "mcafee",
			password: "secret1",
			check:    "secret1",
			x:        6,
			y:        9,
			ferrors: FieldErrors{
				"Password": "password is too short",
			},
		},
	}
	for _, test := range tests {
		// 1. Create form data
		formData := url.Values{}
		formData.Set("username", test.username)
		formData.Set("password", test.password)
		formData.Set("password_check", test.check)
		formData.Set("captcha_answer.X", strconv.Itoa(test.x))
		formData.Set("captcha_answer.Y", strconv.Itoa(test.y))

		// 2. Create a fake request with that data
		req := httptest.NewRequest(
			http.MethodPost,
			"/register",
			strings.NewReader(formData.Encode()), // Encode form data as body
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		cs := &captcha.Solution{
			Radius: 1,
			X:      6,
			Y:      9,
		}

		fh := NewFormHandler[RegisterForm]()
		fd, err := fh.Parse(req, cs)
		require.NoError(t, err)

		if len(test.ferrors) != 0 {
			require.Equal(t, test.ferrors, fd.Errors)
			continue
		}

		require.Equal(t, fd.Data.Username, test.username)
		require.Equal(t, fd.Data.Password, test.password)
		require.Equal(t, fd.Data.Password, test.check)
		require.Equal(t, 6, fd.Data.CaptchaAnswer.X)
		require.Equal(t, 9, fd.Data.CaptchaAnswer.Y)
	}
}

func TestCreateListingForm(t *testing.T) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	// 1. Create form data
	mw.WriteField("title", "blueberry")
	mw.WriteField("description", "these are good blueberries")
	mw.WriteField("category_id", uuid.New().String())
	mw.WriteField("price_tiers.0.quantity", "1")
	mw.WriteField("price_tiers.0.price_cent", "2000")
	mw.WriteField("price_tiers.1.quantity", "2")
	mw.WriteField("price_tiers.1.price_cent", "4000")
	mw.WriteField("captcha_answer.X", strconv.Itoa(6))
	mw.WriteField("captcha_answer.Y", strconv.Itoa(9))

	mwFile, _ := mw.CreateFormFile("image", "product.png")

	img := image.NewNRGBA(image.Rect(0, 0, 1920, 1080))
	for range 1000 {
		x := rand.Int() % 10
		y := rand.Int() % 10
		img.SetNRGBA(x, y, color.NRGBA{R: uint8(rand.Int() % 255)})
	}
	png.Encode(mwFile, img)

	mw.Close()

	// 2. Create a fake request with that data
	req := httptest.NewRequest(http.MethodPost, "/create-listing", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	cs := &captcha.Solution{
		Radius: 1,
		X:      6,
		Y:      9,
	}

	fh := NewFormHandler[CreateListingForm]().WithFileRules(ListingFormFileRules)
	fd, err := fh.Parse(req, cs)
	require.NoError(t, err)
	require.Equal(t, fd.Data.Title, "blueberry")
	require.Equal(t, len(fd.Files), 1)

	file := fd.Files["image"]
	require.Equal(t, "product.png", file.Filename)
	require.Equal(t, "image/png", file.MimeType)
	img2, err := png.Decode(fd.Files["image"].File)
	require.NoError(t, err)
	require.Equal(t, img, img2)
}
