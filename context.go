package tinyweb

import "net/http"

type WebContext struct {
	Request        *http.Request
	ResponseWriter http.ResponseWriter
}
