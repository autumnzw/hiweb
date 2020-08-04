package controllers

import (
	"net/http"

	"github.com/autumnzw/hiweb"
)

type User struct {
	hiweb.Controller
}

// @Description get user
// @Auth
// @httpPost
func (u *User) GetUser() {
	userName := u.GetClaim("name")
	err := u.ServeJSON(http.StatusOK, userName)
	if err != nil {
		u.InternalServerError()
		return
	}
}
