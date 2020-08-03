package hiweb

import (
	"io/ioutil"
	"net/http"
)

type WebContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Body           []byte
}

func (c *WebContext) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *WebContext) GetRemoteAddr() string {
	header := c.Request.Header.Get("X-Forwarded-For")
	if header == "" {
		return c.Request.RemoteAddr
	}
	return header
}

func (c *WebContext) GetBody() ([]byte, error) {
	if len(c.Body) == 0 {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return body, err
		}
		c.Body = body
	}
	return c.Body, nil
}
