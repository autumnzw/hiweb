package hiweb

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func JwtToken(infos map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	for k, v := range infos {
		claims[k] = v
	}
	token.Claims = claims

	return token.SignedString([]byte(WebConfig.SecretKey))
}

var Session *sync.Map

type sessionInfo struct {
	createTime time.Time
	info       map[string]interface{}
}

func InitSession() {
	Session = &sync.Map{}

	go func() {
		for true {
			Session.Range(func(sid, value interface{}) bool {
				if value.(sessionInfo).createTime.Unix()+30*60*1000 < time.Now().Unix() {
					Session.Delete(sid)
				}
				return true
			})
			time.Sleep(time.Duration(3600) * time.Second)
		}
	}()
}

func SessionToken(infos map[string]interface{}) string {
	sid := UUID32()
	Session.Store(sid, sessionInfo{
		createTime: time.Now(),
		info:       infos,
	})
	return sid
}

func SessionDelKey(sid string) {
	Session.Delete(sid)
}

func SessionGetVal(sid string) (map[string]interface{}, bool) {
	if si, has := Session.Load(sid); has {
		return si.(sessionInfo).info, has
	}
	return nil, false
}

func SessionUpdateVal(sid string, info map[string]interface{}) error {
	if sid == "" {
		return fmt.Errorf("sid is blank")
	}
	Session.Store(sid, sessionInfo{
		createTime: time.Now(),
		info:       info,
	})
	return nil
}
