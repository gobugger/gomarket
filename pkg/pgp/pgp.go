package pgp

import (
	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"log/slog"
	"strings"
)

func PublicKeyIsValid(key string) bool {
	if _, err := crypto.NewKeyFromArmored(key); err == nil {
		return true
	} else {
		slog.Debug("invalid pgp key", slog.Any("error", err))
		return false
	}
}

func SignatureIsValid(pubKey string, signature string) bool {
	key, err := crypto.NewKeyFromArmored(pubKey)
	if err != nil {
		return false
	}
	pgp := crypto.PGP()
	verifier, err := pgp.Verify().VerificationKey(key).New()
	if err != nil {
		return false
	}
	verifyResult, err := verifier.VerifyCleartext([]byte(signature))
	if err != nil {
		return false
	}
	if sigErr := verifyResult.SignatureError(); sigErr != nil {
		return false
	}

	return true
}

func Encrypt(pubkey, message string) (string, error) {
	pgp := crypto.PGP()

	publicKey, err := crypto.NewKeyFromArmored(pubkey)
	if err != nil {
		return "", err
	}
	encHandle, err := pgp.Encryption().Recipient(publicKey).New()
	if err != nil {
		return "", err
	}

	encrypted, err := encHandle.Encrypt([]byte(message))
	if err != nil {
		return "", err
	}

	return encrypted.Armor()
}

func Sign(privkey, message string) (string, error) {
	pgp := crypto.PGP()
	privateKey, err := crypto.NewKeyFromArmored(privkey)
	if err != nil {
		return "", err
	}

	signHandle, err := pgp.Sign().SigningKey(privateKey).New()
	if err != nil {
		return "", err
	}

	sign, err := signHandle.SignCleartext([]byte(message))
	if err != nil {
		return "", err
	}

	return string(sign), nil
}

func SignAndEncrypt(privkey, recipientPubKey, message string) (string, error) {
	signed, err := Sign(privkey, message)
	if err != nil {
		return "", err
	}

	return Encrypt(recipientPubKey, signed)
}

func IsEncrypted(msg string) bool {
	msg = strings.Trim(msg, " \n")
	return strings.HasPrefix(msg, "-----BEGIN PGP MESSAGE-----") &&
		strings.HasSuffix(msg, "-----END PGP MESSAGE-----")
}
