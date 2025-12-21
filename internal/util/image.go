package util

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"
)

func DecodeImage(src io.ReadSeeker) (image.Image, error) {
	_, format, err := image.DecodeConfig(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %s", err)
	}

	if _, err := src.Seek(0, 0); err != nil {
		return nil, err
	}

	if format == "webp" {
		return webp.Decode(src)
	} else {
		img, _, err := image.Decode(src)
		return img, err
	}
}

func TransformImage(img image.Image, width, height int) image.Image {
	if width > 0 && height > 0 {
		return imaging.Thumbnail(img, width, height, imaging.Lanczos)
	} else if width > 0 || height > 0 {
		return imaging.Resize(img, width, height, imaging.Lanczos)
	} else {
		return img
	}
}

func SaveImage(ctx context.Context, client *minio.Client, bucketName, objectName string, img image.Image) error {
	buf := bytes.Buffer{}

	if err := webp.Encode(&buf, img, &webp.Options{Lossless: true, Quality: webp.DefaulQuality, Exact: true}); err != nil {
		return err
	}

	_, err := client.PutObject(
		ctx,
		bucketName,
		objectName,
		&buf,
		int64(buf.Len()),
		minio.PutObjectOptions{ContentType: "image/webp"})

	return err
}

// Caller needs to close
func LoadUpload(ctx context.Context, mc *minio.Client, bucketName, objectName string) (io.ReadSeekCloser, error) {
	obj, err := mc.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func LoadImage(ctx context.Context, mc *minio.Client, bucketName, objectName string) (image.Image, error) {
	obj, err := LoadUpload(ctx, mc, bucketName, objectName)
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	return webp.Decode(obj)
}

func RemoveUpload(ctx context.Context, client *minio.Client, bucketName, objectName string) error {
	return client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func Base64Encode(w io.Writer, img image.Image) error {
	encoder := base64.NewEncoder(base64.StdEncoding, w)
	if err := webp.Encode(encoder, img, nil); err != nil {
		return err
	}
	return encoder.Close()
}

func Base64HTMLSrc(img image.Image) (string, error) {
	buf := &strings.Builder{}
	buf.WriteString(`data:image/webp;base64,`)

	if err := Base64Encode(buf, img); err != nil {
		return "", err
	}

	return buf.String(), nil // nolint
}
