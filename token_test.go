package hiweb

import (
	"fmt"
	"testing"
)

func TestJwtToken(t *testing.T) {
	param := make(map[string]interface{})
	param["aa"] = "bb"
	tstr, err := JwtToken(param)
	if err != nil {
		fmt.Println(err)
		return
	}
	m, err := JwtClaims(tstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(m)
}
