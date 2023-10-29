package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	AttNameConnection = "Q09OTkVDVElPTl9BVFRSSUJVVEU"
	AttJwtToken       = "SldUX1RPS0VOX0FUVFJJQlVURQ"
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_."
)

type (
	MyUtils struct {
		Test     string
		GetTest1 func() string
	}
)

func (m *MyUtils) GetTest() string {
	return "20221234"
}

func GetTest2() string {
	return "abcdefg"
}

func (m *MyUtils) GetTime() int64 {
	return time.Now().Unix()
}

func (m *MyUtils) GetTime1Day() int64 {
	return time.Now().Add(24 * time.Hour).Unix()
}

func (m *MyUtils) GetDay(Time int64) string {
	timeSet := time.Unix(Time, 0)
	year, month, day := timeSet.Date()
	return fmt.Sprintf("%s %d %s %d\n", timeSet.Weekday().String()[:3], day, month.String()[:3], year)
}

func (m *MyUtils) GetYear() int {
	year, _, _ := time.Now().Date()
	return year
}

func (m *MyUtils) GetNextSalt() ([]byte, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	bytes, err := strconv.Atoi(os.Getenv("KEY_LENGTH_BYTES"))
	if err != nil {
		return nil, err
	}

	b := make([]byte, bytes)
	_, err = rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
