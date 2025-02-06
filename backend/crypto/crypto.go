package crypto

import (
	"crypto/ed25519"
	"encoding/base64"
)

func LoadEd22519PrivateKeyFromString(value string) (ed25519.PrivateKey, error) {
	key, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return ed25519.PrivateKey{}, err
	}
	return key, nil
}

func StoreEd25519PrivateAsString(value ed25519.PrivateKey) string {
	return base64.RawStdEncoding.EncodeToString(value)
}

func GenerateEd25519StringKey() (string, error) {
	_, rawPrivateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", err
	}
	privateKey := StoreEd25519PrivateAsString(rawPrivateKey)
	if err != nil {
		return "", err
	}
	return privateKey, nil
}
