package utils

import (
	"log"
	"testing"

	"github.com/google/uuid"
)

func TestParseJWT(t *testing.T) {
	test_jwt := "eyJhbGciOiJSUzI1NiIsImtpZCI6ImM4YWI3MTUzMDk3MmJiYTIwYjQ5Zjc4YTA5Yzk4NTJjNDNmZjkxMTgiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI2ODc4OTA1MzA1MS1lcDQ4dDVxbGNrNThnYW0xaGlkNDFnYmJvaHYzbHIyNi5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsImF1ZCI6IjY4Nzg5MDUzMDUxLWVwNDh0NXFsY2s1OGdhbTFoaWQ0MWdiYm9odjNscjI2LmFwcHMuZ29vZ2xldXNlcmNvbnRlbnQuY29tIiwic3ViIjoiMTE2MDY5ODg0OTYzMDA3NTM0NTU3IiwiaGQiOiJrdGhhaXMuY29tIiwiZW1haWwiOiJ2aXZpZW5uZUBrdGhhaXMuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImF0X2hhc2giOiJULUlrQXdROHVzRXFFLUFhOU5oT2JRIiwibmFtZSI6IlZpdmllbm5lIEN1cmV3aXR6IiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hL0FDZzhvY0t1bURQRTZoWW5ERzY4bkt0Nlp1Q3JtbGtMUEJPbXpxQVBValp3bFJGTWpYU3ZBZz1zOTYtYyIsImdpdmVuX25hbWUiOiJWaXZpZW5uZSIsImZhbWlseV9uYW1lIjoiQ3VyZXdpdHoiLCJpYXQiOjE3NTk5NDY2ODEsImV4cCI6MTc1OTk1MDI4MX0.afkVhYSvqRvPjnkfNOmAkiS9xo-mwFKy2cgzR2xndHRE1XE8DrVuHZflIdG3PMknW4-R6tfhHZZqLH7OXJGgg9AKbQk1C_pl10rk_5hG5vm2b2Es_r32M6ZyUEHIQqBeT5AQ4cSTPmduQuTGwI9Ku0dv12xJQ25FJXmzktHe0ijMNBeIeScfhXZqlQSs6Dgutd2n9YaJCJmcxFbFihN8fyhXEXWm6F4BdjtE6bY2oPUKTxrlbelkz-B0z72Y6utZrBnDMEFmfpARISmUynaCM5zVA4DX4IeVtTGRPuAG-Y2T0jAM-s6wn4nGXZYbuRDW6Q7qVz-gX_1Pv9e63NZJmQ"
	_, err := ParseJWT(string(test_jwt))
	if err != nil {
		t.Errorf("Error Parsing JWT %v\n", err)
	}
}

func TestKeyFetch(t *testing.T) {
	GetGoogleJWKSKey()
}

<<<<<<< HEAD
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
=======
func TestParseAndVerifyGoogle(t *testing.T) {
	filename := "google.jwt"
	f, _ := os.Open(filename)
	reader := bufio.NewReader(f)
	test_jwt, isPrefix, _ := reader.ReadLine()
	if isPrefix {
		log.Println("Long JWT")
	}
	valid, token := ParseAndVerifyGoogle(string(test_jwt))
>>>>>>> looks like jwt pipeline is working
	if !valid {
		t.Errorf("Could not validate JWT: \n")
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
	uuid, _ := uuid.Parse("50c06e4d-b594-4489-9d4b-a513f63c90bd")
	newJwt := WriteJWT("vivienne@kthais.com", []string{"user", "admin", "queen"}, uuid, key, 15)
	log.Printf("JWT Generated: %v\n", newJwt)
}
