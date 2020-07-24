package hiweb

import "net/http"

type WebContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
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
