package controllers

import (
	// "caoweb/model"

	"net/http"
	"time"

	"github.com/autumnzw/hiweb"
)

type Token struct {
	hiweb.Controller
}
type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//@httpPost
func (t *Token) GenToken(userIn UserCredentials) {
	if userIn.Username != "admin" || userIn.Password != "admin" {
		t.Forbidden()
		return
	}
	expiresAt := time.Now().Add(time.Hour * time.Duration(1)).Unix()
	authTime := time.Now().Unix()
	tokenParam := map[string]interface{}{
		"name": userIn.Username,
		"exp":  expiresAt,
		"iat":  authTime,
	}
	tokenString, err := hiweb.JwtToken(tokenParam)
	if err != nil {
		t.Forbidden()
		return
	}
	_ = t.ServeJSON(http.StatusOK, struct {
		Token     string `json:"token"`
		TokenType string `json:"tokenType"`
		Profile   struct {
			Key       string `json:"key"`
			Name      string `json:"name"`
			AuthTime  int64  `json:"authTime"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"profile"`
	}{
		TokenType: "Bearer",
		Token:     tokenString,
		Profile: struct {
			Key       string `json:"key"`
			Name      string `json:"name"`
			AuthTime  int64  `json:"authTime"`
			ExpiresAt int64  `json:"expiresAt"`
		}{
			Name:      userIn.Username,
			ExpiresAt: expiresAt,
			AuthTime:  authTime,
		},
	})
}
