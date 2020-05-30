package main

import (
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"net/http"
	"time"
)

type CustomClaims struct {
	*jwt.Claims
	Authorized bool   `json:"authorized,omitempty"`
	UserId     int    `json:"user_id,omitempty"`
	UserName   string `json:"user_name,omitempty"`
	UserEmail  string `json:"user_email,omitempty"`
}

const JWTLifeTime = time.Minute * 5

var rsaSigner jose.Signer
var jwk jose.JSONWebKey

func initJWKsEndpoint(router *gin.Engine) {
	router.GET("/.well-known/jwks.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"keys": []jose.JSONWebKey{jwk}})
	})
}

func initJWTSecrets() error {
	var err error
	pemData := viper.GetString("secret")
	block, _ := pem.Decode([]byte(pemData))
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	key := jose.SigningKey{Algorithm: jose.RS256, Key: privKey}
	var signerOpts = jose.SignerOptions{}
	signerOpts.WithType("JWT")
	signerOpts.WithHeader("kid", "1")

	rsaSigner, err = jose.NewSigner(key, &signerOpts)
	if err != nil {
		return err
	}
	jwk = jose.JSONWebKey{
		Key:                         privKey.Public(),
		KeyID:                       "1",
		Algorithm:                   "RSA",
	}
	return nil
}

func createJWT(user *User) (string, error) {
	var err error

	builder := jwt.Signed(rsaSigner)
	claims := CustomClaims{
		Claims: &jwt.Claims{
			Issuer:   "goauth",
			Subject:  "userAuth",
			ID:		  "1",
			Audience: jwt.Audience{"aud1", "aud2"},
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
			Expiry:   jwt.NewNumericDate(time.Now().UTC().Add(JWTLifeTime)),
		},
		Authorized: true,
		UserId:     user.Id,
		UserName:   user.Name,
		UserEmail:  user.Email,
	}
	builder = builder.Claims(claims)

	token, err := builder.CompactSerialize()
	if err != nil {
		return "", err
	}

	return token, nil
}
