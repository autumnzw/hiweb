package tinyweb

import (
	"github.com/dgrijalva/jwt-go"
)

func GenToken(infos map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	for k, v := range infos {
		claims[k] = v
	}
	token.Claims = claims

	return token.SignedString([]byte(WebConfig.SecretKey))
}
