package tinyweb

import (
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"sync"
	"webclient/tinyweb/swaggerFiles"
)

// WrapHandler wraps swaggerFiles.Handler and returns http.HandlerFunc
var WrapHandler = Handler()

// Config stores httpSwagger configuration variables.
type SwaggerConfig struct {
	//The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	ProjectName  string
	URL          string
	DeepLinking  bool
	DocExpansion string
	DomID        string
}

// URL presents the url pointing to API definition (normally swagger.json or swagger.yaml).
func URL(url string, projectName string) func(c *SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.URL = url
		c.ProjectName = projectName
	}
}

// DeepLinking true, false
func DeepLinking(deepLinking bool) func(c *SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DeepLinking = deepLinking
	}
}

// DocExpansion list, full, none
func DocExpansion(docExpansion string) func(c *SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DocExpansion = docExpansion
	}
}

// DomID #swagger-ui
func DomID(domID string) func(c *SwaggerConfig) {
	return func(c *SwaggerConfig) {
		c.DomID = domID
	}
}

// Handler wraps `http.Handler` into `http.HandlerFunc`.
func Handler(configFns ...func(*SwaggerConfig)) http.HandlerFunc {
	config := &SwaggerConfig{
		URL:          "doc.json",
		DeepLinking:  true,
		DocExpansion: "list",
		DomID:        "#swagger-ui",
		ProjectName:  "",
	}
	for _, configFn := range configFns {
		configFn(config)
	}

	//create a template with name
	t := template.New("swagger_index.html")
	index, _ := t.Parse(indexTempl)

	var re = regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	return func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Headers", "*")
		headers.Set("Access-Control-Allow-Method", "*")
		headers.Set("Access-Control-Allow-Credentials", "true")

		matches := re.FindStringSubmatch(r.RequestURI)
		path := matches[2]
		prefix := matches[1]

		h := swaggerFiles.Handler
		h.Prefix = prefix

		switch path {
		case "index.html":
			_ = index.Execute(w, config)
		case "swagger.json":
			doc, err := SwaggerReadDoc()
			if err != nil {
				panic(err)
			}
			_, _ = w.Write([]byte(doc))
		case "":
			http.Redirect(w, r, prefix+"index.html", 301)
		default:
			h.ServeHTTP(w, r)
		}
		return
	}
}

const indexTempl = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
      box-sizing: inherit;
    }

    body {
      margin:0;
      background: #fafafa;
    }
  </style>
  
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
          <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>


    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>

  </defs>
</svg>

<div id="swagger-ui"></div>

<!-- Workaround for https://github.com/swagger-api/swagger-editor/issues/1371 -->
<script>
  if (window.navigator.userAgent.indexOf("Edge") > -1) {
    console.log("Removing native Edge fetch in favor of swagger-ui's polyfill")
    window.fetch = undefined;
  }
</script>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
  window.onload = function () {
    var configObject = JSON.parse('{"urls":[{"url":"/swag/swagger.json","name":"{{.ProjectName}}"}],"deepLinking":false,"displayOperationId":false,"defaultModelsExpandDepth":1,"defaultModelExpandDepth":1,"defaultModelRendering":"example","displayRequestDuration":false,"docExpansion":"list","showExtensions":false,"showCommonExtensions":false,"supportedSubmitMethods":["get","put","post","delete","options","head","patch","trace"],"validatorUrl":null}');
    var oauthConfigObject = JSON.parse('{"clientId":"clientId","clientSecret":"clientSecret","scopeSeperator":" ","useBasicAuthenticationWithAccessCodeGrant":false}');

    // Apply mandatory parameters
    configObject.dom_id = "#swagger-ui";
    configObject.presets = [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset];
    configObject.layout = "StandaloneLayout";

    // If oauth2RedirectUrl isn't specified, use the built-in default
    if (!configObject.hasOwnProperty("oauth2RedirectUrl"))
      configObject.oauth2RedirectUrl = window.location.href.replace("index.html", "oauth2-redirect.html").split('#')[0];

    // Build a system
    const ui = SwaggerUIBundle(configObject);

    // Apply OAuth config
    ui.initOAuth(oauthConfigObject);
  }
</script>
</body>

</html>
`

var (
	swaggerMu sync.RWMutex
	swag      Swagger
)

// Swagger is a interface to read swagger document.
type Swagger interface {
	ReadDoc() string
}

// Register registers swagger for given name.
func SwaggerRegister(swagger Swagger) {
	swaggerMu.Lock()
	defer swaggerMu.Unlock()
	if swagger == nil {
		panic("swagger is nil")
	}

	if swag != nil {
		panic("Register called twice for swag")
	}
	swag = swagger
}

// ReadDoc reads swagger document.
func SwaggerReadDoc() (string, error) {
	if swag != nil {
		return swag.ReadDoc(), nil
	}
	return "", errors.New("not yet registered swag")
}
