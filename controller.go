package hiweb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"

	"github.com/go-playground/validator/v10"
)

type ControllerInterface interface {
	Init(ctx *WebContext)
	CheckAuth() (bool, error)
	Query(key string, def ...interface{}) (interface{}, error)
	ParseValid(obj interface{}, vs ...*validator.Validate) error
}
type Controller struct {
	// context data
	Ctx       *WebContext
	Claims    jwt.MapClaims
	JsonParam map[string]interface{}
}

func (c *Controller) SetHeader(key, val string) {
	c.Ctx.ResponseWriter.Header().Set(key, val)
}

func (c *Controller) GetBody() ([]byte, error) {
	return c.Ctx.GetBody()
}

func (c *Controller) Input() url.Values {
	if c.Ctx.Request.Form == nil {
		err := c.Ctx.Request.ParseForm()
		if err != nil {
			WebConfig.Logger.Error(err)
		}
	}
	return c.Ctx.Request.Form
}

func (c *Controller) Init(ctx *WebContext) {
	c.Ctx = ctx
}

func (c *Controller) CheckAuth() (bool, error) {
	token, err := request.ParseFromRequest(c.Ctx.Request, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(WebConfig.SecretKey), nil
		})
	if err == nil {
		if token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Claims = claims
			}
			return true, nil
		} else {
			return false, fmt.Errorf("Token is not valid")
		}
	} else {
		return false, fmt.Errorf("Unauthorized access to this resource")
	}
}

func (c *Controller) GetClaim(key string) interface{} {
	if v, h := c.Claims[key]; h {
		return v
	} else {
		return nil
	}
}

// ParseValid maps input data map to obj struct.include(form,json)
func (c *Controller) ParseValid(obj interface{}, vs ...*validator.Validate) error {
	contentType := c.Ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/*+json") {
		tBody, err := c.GetBody()
		if err != nil {
			return err
		}
		requestBody := bytes.TrimSpace(tBody)
		if len(requestBody) != 0 && IsJSONBody(requestBody) {
			err = json.Unmarshal(requestBody, obj)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("input not json")
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := c.Ctx.Request.ParseMultipartForm(1 << 26)
		if err != nil {
			return err
		}
		err = ParseForm(c.Input(), obj)
		if err != nil {
			return err
		}
	} else {
		err := ParseForm(c.Input(), obj)
		if err != nil {
			return err
		}
	}
	if len(vs) > 0 {
		for _, v := range vs {
			err := v.Struct(obj)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Param returns router param by a given key.
func (c *Controller) Param(key string) string {
	if v, has := c.Input()[key]; has {
		return v[0]
	}
	return ""
}

func (c *Controller) ParseJson() error {
	if c.JsonParam == nil {
		c.JsonParam = make(map[string]interface{})
		tBody, err := c.GetBody()
		if err != nil {
			return err
		}
		requestBody := bytes.TrimSpace(tBody)
		if len(requestBody) != 0 && IsJSONBody(requestBody) {
			obj := make(map[string]interface{})
			err := json.Unmarshal(requestBody, &obj)
			if err != nil {
				return err
			}
			for k, v := range obj {
				c.JsonParam[k] = v
			}
		}
	}
	return nil
}

// Param returns router param by a given key.
func (c *Controller) Query(key string, def ...interface{}) (interface{}, error) {
	if v := c.Param(key); v != "" {
		return v, nil
	}
	if len(def) > 0 {
		return def[0], nil
	}

	contentType := c.Ctx.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "application/json") ||
		strings.HasPrefix(contentType, "application/*+json") {
		err := c.ParseJson()
		if err != nil {
			return "", err
		}
		if v, has := c.JsonParam[key]; has {
			return v, nil
		} else {
			return "", fmt.Errorf("not found:%s", key)
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		err := c.Ctx.Request.ParseMultipartForm(1 << 26)
		if err != nil {
			return "", err
		}
		if v := c.Param(key); v != "" {
			return v, nil
		}
		return "", nil
	} else if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		tBody, err := c.GetBody()
		if err != nil {
			return "", err
		}
		requestBody := bytes.TrimSpace(tBody)
		if len(requestBody) != 0 {
			u, err := url.ParseQuery(string(requestBody))
			if err != nil {
				return "", nil
			}
			if v, has := u[key]; has {
				return v[0], nil
			} else {
				return "", nil
			}
		}
	}
	return "", nil
}

// GetString returns the input value by key string or the default value while it's present and input is blank
func (c *Controller) GetString(key string, def ...string) (string, error) {
	val, err := c.Query(key, def)
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// GetStrings returns the input string slice by key string or the default value while it's present and input is blank
// it's designed for multi-value input field such as checkbox(input[type=checkbox]), multi-selection.
func (c *Controller) GetStrings(key string, def ...[]string) ([]string, error) {
	val, err := c.Query(key, def)
	if err != nil {
		return []string{}, err
	}
	return val.([]string), nil
}

// GetInt returns input as an int or the default value while it's present and input is blank
func (c *Controller) GetInt(key string, def ...int) (int, error) {
	val, err := c.Query(key, def)
	if err != nil {
		return -1, err
	}
	return val.(int), nil

}

// ParseCheck maps input data map to obj struct.include(form,json)
func (c *Controller) ParseCheck(obj interface{}) error {
	valid := validator.New()
	err := c.ParseValid(obj, valid)
	return err
}

func (c *Controller) Forbidden() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
}
func (c *Controller) InternalServerError() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
}
func (c *Controller) NotFound() {
	c.Ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
}

func (c *Controller) ServeJSON(status int, obj interface{}) error {
	return c.JSON(status, obj, true, false)
}

func (c *Controller) JSON(status int, data interface{}, hasIndent bool, coding bool) error {
	c.SetHeader("Content-Type", "application/json; charset=utf-8")
	var content []byte
	var err error
	if hasIndent {
		content, err = json.MarshalIndent(data, "", "  ")
	} else {
		content, err = json.Marshal(data)
	}
	if err != nil {
		http.Error(c.Ctx.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return err
	}
	if coding {
		content = []byte(stringsToJSON(string(content)))
	}
	return c.ServeBody(status, content)
}

func (c *Controller) ServeBody(status int, content []byte) error {
	var encoding string
	var buf = &bytes.Buffer{}
	if WebConfig.EnableGzip {
		encoding = ParseEncoding(c.Ctx.Request)
	}
	if b, n, _ := WriteBody(encoding, buf, content); b {
		c.SetHeader("Content-Encoding", n)
		c.SetHeader("Content-Length", strconv.Itoa(buf.Len()))
	} else {
		c.SetHeader("Content-Length", strconv.Itoa(len(content)))
	}
	// Write status code if it has been set manually
	// Set it to 0 afterwards to prevent "multiple response.WriteHeader calls"
	if status != 0 {
		c.Ctx.ResponseWriter.WriteHeader(status)
	}
	_, err := io.Copy(c.Ctx.ResponseWriter, buf)
	return err
}

//ServeDownload
func (c *Controller) ServeDownload(file string, filename ...string) {
	// check get file error, file not found or other error.
	if _, err := os.Stat(file); err != nil {
		http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, file)
		return
	}

	var fName string
	if len(filename) > 0 && filename[0] != "" {
		fName = filename[0]
	} else {
		fName = filepath.Base(file)
	}
	c.SetHeader("Content-Disposition", "attachment; filename="+url.QueryEscape(fName))
	c.SetHeader("Content-Description", "File Transfer")
	c.SetHeader("Content-Type", "application/octet-stream")
	c.SetHeader("Content-Transfer-Encoding", "binary")
	c.SetHeader("Expires", "0")
	c.SetHeader("Cache-Control", "must-revalidate")
	c.SetHeader("Pragma", "public")
	http.ServeFile(c.Ctx.ResponseWriter, c.Ctx.Request, file)
}

// ServeDownloadContent下载文件
func (c *Controller) ServeDownloadContent(status int, content []byte, fileName string) error {
	c.SetHeader("Content-Disposition", "attachment; filename="+url.QueryEscape(fileName))
	c.SetHeader("Content-Description", "File Transfer")
	c.SetHeader("Content-Type", "application/octet-stream")
	c.SetHeader("Content-Transfer-Encoding", "binary")
	c.SetHeader("Expires", "0")
	c.SetHeader("Cache-Control", "must-revalidate")
	c.SetHeader("Pragma", "public")
	return c.ServeBody(status, content)
}

func stringsToJSON(str string) string {
	var jsons bytes.Buffer
	for _, r := range str {
		rint := int(r)
		if rint < 128 {
			jsons.WriteRune(r)
		} else {
			jsons.WriteString("\\u")
			jsons.WriteString(strconv.FormatInt(int64(rint), 16))
		}
	}
	return jsons.String()
}
