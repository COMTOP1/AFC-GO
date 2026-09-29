package utils

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

type (
	Type       int
	Length     int
	Characters string
)

const (
	GeneratePassword Type = iota
	GenerateSalt
	GenerateUsername
)

const (
	UsernameCharacters Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	SaltCharacters     Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789/."
	PasswordCharacters Characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@*()&"
)

const (
	PasswordLength Length = 12
	SaltLength     Length = 22
	UsernameLength Length = 20
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

func GenerateRandomLength(length int, randomType Type) (string, error) {
	if length < 6 || length > 40 {
		return "", errors.New("length must be between 6 and 40")
	}
	switch randomType {
	case GeneratePassword:
		b, err := rangeLoop(PasswordCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random password: %w", err)
		}

		return b, nil
	case GenerateSalt:
		b, err := rangeLoop(SaltCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random salt: %w", err)
		}

		return hex.EncodeToString(sha512.New().Sum([]byte(b))), nil
	case GenerateUsername:
		b, err := rangeLoop(UsernameCharacters, Length(length))
		if err != nil {
			return "", fmt.Errorf("error generating random username: %w", err)
		}

		return b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", randomType)
	}
}

// GenerateRandom generates a random string for either password or salt
func GenerateRandom(randomType Type) (string, error) {
	switch randomType {
	case GeneratePassword:
		b, err := rangeLoop(PasswordCharacters, PasswordLength)
		if err != nil {
			return "", fmt.Errorf("error generating random password: %w", err)
		}

		return b, nil
	case GenerateSalt:
		b, err := rangeLoop(SaltCharacters, SaltLength)
		if err != nil {
			return "", fmt.Errorf("error generating random salt: %w", err)
		}

		return hex.EncodeToString(sha512.New().Sum([]byte(b))), nil
	case GenerateUsername:
		b, err := rangeLoop(UsernameCharacters, UsernameLength)
		if err != nil {
			return "", fmt.Errorf("error generating random username: %w", err)
		}

		return b, nil
	default:
		return "", fmt.Errorf("invalid type: %d", randomType)
	}
}

// rangeLoop creates the random string that will be used
func rangeLoop(characters Characters, randomLength Length) (string, error) {
	bytes := make([]byte, randomLength)

	length := big.NewInt(int64(len(characters)))

	for i := range bytes {
		randInt, err := rand.Int(rand.Reader, length)
		if err != nil {
			return "", err
		}

		bytes[i] = characters[randInt.Int64()]
	}

	return string(bytes), nil
}
