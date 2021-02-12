package test

import (
	"math/rand"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func GenerateName() string {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
