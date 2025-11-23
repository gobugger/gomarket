package qrcode

import (
	"encoding/base64"
	qr "github.com/skip2/go-qrcode"
	"io"
	// "bytes"
	// qrterminal "github.com/mdp/qrterminal/v3"
)

func Base64Encode(dst io.Writer, content string) error {
	base64Encoder := base64.NewEncoder(base64.StdEncoding, dst)
	defer base64Encoder.Close()

	q, err := qr.New(content, qr.Medium)
	if err != nil {
		return err
	}

	return q.Write(164, base64Encoder)
}

/*
func AsString(content string) string {
	buf := bytes.NewBuffer(make([]byte, 0))
	qrterminal.Generate(content, qrterminal.L, buf)
	data, _ := io.ReadAll(buf)
	return string(data)
}
*/
