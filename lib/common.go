package lib

import (
	"log"
	"math/rand"
	"os"
)

var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// FailOnError multi func for error HIGH
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// LogOnError log error not panic program
func LogOnError(err error, msg string) bool {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		return false
	}

	return true
}

// CheckFileExist check file on disk
func CheckFileExist(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}

	return nil
}

// RandStringRunes generate string
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
