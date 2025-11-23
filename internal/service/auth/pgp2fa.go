package auth

import (
	"crypto/rand"
	"fmt"
	"github.com/gobugger/gomarket/pkg/pgp"
)

const TokenLength = 26

type Challenge struct {
	Token            string
	EncryptedMessage string
}

func Generate2FAChallenge(pgpKey string) (*Challenge, error) {
	token := rand.Text()

	message := fmt.Sprintf("Code: %s\n", token)
	encryptedMessage, err := pgp.Encrypt(pgpKey, message)
	if err != nil {
		return nil, err
	}

	return &Challenge{
		Token:            token,
		EncryptedMessage: encryptedMessage,
	}, nil
}
