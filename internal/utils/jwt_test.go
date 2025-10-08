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

func TestParseAndVerifyGoogle(t *testing.T) {
	filename := "google.jwt"
	f, _ := os.Open(filename)
	reader := bufio.NewReader(f)
	test_jwt, isPrefix, _ := reader.ReadLine()
	if isPrefix {
		log.Println("Long JWT")
	}
	valid, token := ParseAndVerifyGoogle(string(test_jwt))
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

func TestParseAndVerify(t *testing.T) {
	key := "testkey123456"
	newJwt := WriteJWT("vivienne@kthais.com", []string{"user", "admin", "queen"}, key, 15)
	valid, _ := ParseAndVerify(newJwt, key)
	if !valid {
		t.Errorf("Could not validate JWT: \n")
	}
}

func TestJWTCreate(t *testing.T) {
	key := "testkey123456"
	newJwt := WriteJWT("vivienne@kthais.com", []string{"user", "admin", "queen"}, key, 15)
	log.Printf("JWT Generated: %v\n", newJwt)
}
