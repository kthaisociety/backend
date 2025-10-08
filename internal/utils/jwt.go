package utils

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// this module will hanlde confirming google logins and issueing our own jwt, which will have user
// info attached.
// It will also decode jwts and validate them. Most auth pipeline steps should use this.

func ParseJWT(jwtIn string) (*jwt.Token, error) {
	// options := [jwt.WithoutClaimsValidtion()]
	jwtParser := jwt.NewParser()
	token, _, err := jwtParser.ParseUnverified(jwtIn, jwt.MapClaims{})
	return token, err
}

func getKeyById(kl *KeyList, id string) *JwksKey {
	for _, key := range kl.Keys {
		if key.Kid == id {
			return &key
		}
	}
	return nil
}

func keyfunc(token *jwt.Token) (any, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("token has no kid")
	}

	klist := GetGoogleJWKSKey()

	key := getKeyById(klist, kid)
	if key == nil {
		return nil, fmt.Errorf("no matching key found")
	}

	// Decode n and e (base64url to big.Int) and build rsa.PublicKey
	nBytes, _ := base64.RawURLEncoding.DecodeString(key.N)
	eBytes, _ := base64.RawURLEncoding.DecodeString(key.E)

	n := new(big.Int).SetBytes(nBytes)
	// e is usually 65537
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

func ParseAndVerify(jwtIn string) (bool, *jwt.Token) {
	jwtParser := jwt.NewParser()

	token, err := jwtParser.Parse(jwtIn, keyfunc)
	if err != nil {
		log.Printf("Error parsing encrypted token")
	}
	return token.Valid, token
}

func GetClaims(token *jwt.Token) jwt.MapClaims {
	claims, _ := token.Claims.(jwt.MapClaims)
	return claims

}

type JwksKey struct {
	N   string `json:"n"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	E   string `json:"e"`
}

type KeyList struct {
	Keys []JwksKey `json:"keys"`
}

func GetGoogleJWKSKey() *KeyList {
	url := "https://www.googleapis.com/oauth2/v3/certs"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Could not retrieve google credentials")
		return nil
	}
	defer resp.Body.Close()
	var data KeyList
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Error decoding json%v\n", err)
	}
	return &data
}
