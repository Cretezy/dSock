package common

import (
	"math/rand"
	"time"
)

func RemoveString(texts []string, text string) []string {
	for index, value := range texts {
		if value == text {
			return append(texts[:index], texts[index+1:]...)
		}
	}

	return texts
}

func UniqueString(texts []string) []string {
	keys := make(map[string]struct{})

	removed := 0
	for _, text := range texts {
		if _, value := keys[text]; !value {
			keys[text] = struct{}{}
			texts[removed] = text
			removed++
		}
	}
	return texts[:removed]
}

func RemoveEmpty(texts []string) []string {
	removed := 0
	for _, text := range texts {
		if text != "" {
			texts[removed] = text
			removed++
		}
	}
	return texts[:removed]
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(length int) string {
	random := make([]byte, length)

	for index := range random {
		random[index] = charset[seededRand.Intn(len(charset))]
	}

	return string(random)
}
