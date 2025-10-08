package utils

import (
	"bufio"
	"log"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseJWT(t *testing.T) {
	filename := "google.jwt"
	f, _ := os.Open(filename)
	reader := bufio.NewReader(f)
	test_jwt, isPrefix, _ := reader.ReadLine()
	if isPrefix {
		log.Println("Long JWT")
	}
	_, err := ParseJWT(string(test_jwt))
	if err != nil {
		t.Errorf("Error Parsing JWT %v\n", err)
	}
}

func TestKeyFetch(t *testing.T) {
	GetGoogleJWKSKey()
}

func TestParseAndVerify(t *testing.T) {
	filename := "google.jwt"
	f, _ := os.Open(filename)
	reader := bufio.NewReader(f)
	test_jwt, isPrefix, _ := reader.ReadLine()
	if isPrefix {
		log.Println("Long JWT")
	}
	valid, token := ParseAndVerify(string(test_jwt))
	if !valid {
		t.Errorf("Token not valid!!")
	} else {
		claims, _ := token.Claims.(jwt.MapClaims)
		for keys, values := range claims {
			log.Printf("Key: %v --- Value: %v\n", keys, values)
		}
	}
	if false {
		t.Errorf("No error\n")
	}
}
