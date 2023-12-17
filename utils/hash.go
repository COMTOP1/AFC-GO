package utils

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

type Type int

const (
	GeneratePassword Type = iota
	GenerateSalt
)

const (
	saltCharacters     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/."
	passwordCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@*()&"
)

func HashPass(password, salt []byte, iter, keyLen int) []byte {
	return pbkdf2.Key(password, salt, iter, keyLen, sha512.New)
}

func HashPassScrypt(password, salt []byte, workFactor, blockSize, parallelismFactor, keyLen int) (string, error) {
	hash, err := scrypt.Key(password, salt, workFactor, blockSize, parallelismFactor, keyLen)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}
	return hex.EncodeToString(hash), nil
}

// GenerateRandom generates a random string for either password or salt
func GenerateRandom(t Type) (string, error) {
	switch t {
	case GeneratePassword:
		lenPass := big.NewInt(int64(len(passwordCharacters)))
		b, err := rangeLoop(lenPass, 12)
		if err != nil {
			return "", fmt.Errorf("error generating random: %w", err)
		}
		return b, nil
	case GenerateSalt:
		lenSalt := big.NewInt(int64(len(saltCharacters)))
		b, err := rangeLoop(lenSalt, 32)
		if err != nil {
			return "", fmt.Errorf("error generating random: %w", err)
		}
		return b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", t)
	}
}

// rangeLoop creates the random string that will be used
func rangeLoop(len *big.Int, size int) (string, error) {
	b := make([]byte, size)
	for i := range b {
		randInt, err := rand.Int(rand.Reader, len)
		if err != nil {
			return "", err
		}
		b[i] = passwordCharacters[randInt.Int64()]
	}
	return string(b), nil
}
