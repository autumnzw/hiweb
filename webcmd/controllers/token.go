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
func (t *Token) Login(userIn UserCredentials) {

}

//@httpPost /Auth/Login
func (t *Token) GenToken(userIn UserCredentials) {

}

//@httpPost /Auth/Login
func (t *Token) Same(userIn UserCredentials) {

}
