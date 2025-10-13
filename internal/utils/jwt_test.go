package utils

import (
	"bufio"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
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

// this is only to confirm that we can verify a google token, but requires a fresh token
// func TestParseAndVerifyGoogle(t *testing.T) {
// 	filename := "google.jwt"
// 	f, _ := os.Open(filename)
// 	reader := bufio.NewReader(f)
// 	test_jwt, isPrefix, _ := reader.ReadLine()
// 	if isPrefix {
// 		log.Println("Long JWT")
// 	}
// 	valid, token := ParseAndVerifyGoogle(string(test_jwt))
// 	if !valid {
// 		t.Errorf("Token not valid!!")
// 	} else {
// 		claims, _ := token.Claims.(jwt.MapClaims)
// 		for keys, values := range claims {
// 			log.Printf("Key: %v --- Value: %v\n", keys, values)
// 		}
// 	}
// 	if false {
// 		t.Errorf("No error\n")
// 	}
// }

func TestParseAndVerify(t *testing.T) {
	key := "testkey123456"
	uuid, _ := uuid.Parse("50c06e4d-b594-4489-9d4b-a513f63c90bd")
	newJwt := WriteJWT("vivienne@kthais.com", []string{"user", "admin", "queen"}, uuid, key, 15)
	valid, _ := ParseAndVerify(newJwt, key)
	if !valid {
		t.Errorf("Could not validate JWT: \n")
	}
}

func TestJWTCreate(t *testing.T) {
	key := "testkey123456"
	uuid, _ := uuid.Parse("50c06e4d-b594-4489-9d4b-a513f63c90bd")
	newJwt := WriteJWT("vivienne@kthais.com", []string{"user", "admin", "queen"}, uuid, key, 15)
	log.Printf("JWT Generated: %v\n", newJwt)
}
