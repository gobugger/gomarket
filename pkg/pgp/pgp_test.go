package pgp

import (
	"testing"
)

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		msg    string
		result bool
	}{
		{
			msg:    "-----BEGIN PGP MESSAGE-----\nencrypted data here\n-----END PGP MESSAGE-----",
			result: true,
		},
		{
			msg:    " \n -----BEGIN PGP MESSAGE-----\nencrypted data here\n-----END PGP MESSAGE-----  \n ",
			result: true,
		},
		{
			msg:    "-----BEGIN PGP MESSAGE----\nencrypted data here\n-----END PGP MESSAGE-----",
			result: false,
		},
		{
			msg:    "-----BEGIN PGP MESSAGE-----\nencrypted data here\n-----BEGIN PGP MESSAGE-----",
			result: false,
		},

		{
			msg:    "-----BEGIN PGP MESSAGE-----\nencrypted data here\n",
			result: false,
		},
		{
			msg:    "nencrypted data here\n-----BEGIN PGP MESSAGE-----",
			result: false,
		},
	}

	for _, test := range tests {
		if IsEncrypted(test.msg) != test.result {
			t.Fatalf("should return %v for %s", test.result, test.msg)
		}
	}
}
