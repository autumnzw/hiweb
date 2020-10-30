package controllers

import (
	"github.com/autumnzw/hiweb"
)

type Token struct {
	hiweb.Controller
}
type UserCredentials struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

//@httpPost
func (t *Token) Login(userIn UserCredentials) {

}

//@httpGet
func (t *Token) Get(key string) {

}

//@httpPost /Service/Auth/Login
func (t *Token) GenToken(userIn UserCredentials) {

}

//@httpGetPost /Auth/Login
func (t *Token) Same(userIn UserCredentials) {

}

//@upload file aa
func (t *Token) Upload() {

}
