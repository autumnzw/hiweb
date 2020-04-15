package controllers

import (
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

}
