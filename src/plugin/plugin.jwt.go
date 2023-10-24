package zgwit_plugin

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Info  interface{}
	Timer int64

	jwt.StandardClaims
}

// ReleaseToken 申请令牌
func ReleaseToken(jwtKey []byte, expire int64, info interface{}) (token string, err error) {

	claims := &Claims{
		Info:  info,
		Timer: int64(time.Now().Unix()) + expire*24*60*60,
	}

	_token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err = _token.SignedString(jwtKey)

	return
}

// ParseToken 验证令牌
func ParseToken(jwtKey []byte, tokenString string) (*jwt.Token, int64, interface{}, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtKey, nil
	})

	return token, claims.Timer, claims.Info, err
}
