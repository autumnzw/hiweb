package webcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
)

// Gen presents a generate tool for swag.
type Gen struct {
	jsonIndent func(data interface{}) ([]byte, error)
	jsonToYAML func(data []byte) ([]byte, error)
}

// New creates a new Gen.
func NewGen() *Gen {
	return &Gen{
		jsonIndent: func(data interface{}) ([]byte, error) {
			bf := bytes.NewBuffer([]byte{})
			jsonEncoder := json.NewEncoder(bf)
			jsonEncoder.SetEscapeHTML(false)
			jsonEncoder.SetIndent("", "    ")
			err := jsonEncoder.Encode(data)
			return bf.Bytes(), err
		},
		jsonToYAML: yaml.JSONToYAML,
	}
}

// Config presents Gen configurations.
type Config struct {
	ProjectName string
	// SearchDir the swag would be parse
	SearchDir string

	// OutputDir represents the output directory for all the generated files
	OutputDir string

	VueBaseUrl string

	VueOutputDir string

	// MainAPIFile the Go file path in which 'swagger general API Info' is written
	MainAPIFile string

	// PropNamingStrategy represents property naming strategy like snakecase,camelcase,pascalcase
	PropNamingStrategy string

	// ParseVendor whether swag should be parse vendor folder
	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// MarkdownFilesDir used to find markdownfiles, which can be used for tag descriptions
	MarkdownFilesDir string

	// GeneratedTime whether swag should generate the timestamp at the top of docs.go
	GeneratedTime bool
}

type SwaggerSpec struct {
	OpenApi    string                              `json:"openapi"`
	Info       SwaggerInfo                         `json:"info"`
	Paths      map[string]map[string]SwaggerMethod `json:"paths"`
	Components *SwaggerComponent                   `json:"components"`
}
type SwaggerComponent struct {
	Schema          map[string]SwaggerComponentStruct          `json:"schemas"`
	SecuritySchemes map[string]SwaggerComponentSecuritySchemes `json:"securitySchemes,omitempty"`
}

type SwaggerComponentStruct struct {
	Type                 string                   `json:"type"`
	Properties           map[string]SwaggerSchema `json:"properties"`
	AdditionalProperties bool                     `json:"additionalProperties"`
}

type SwaggerComponentSecuritySchemes struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Name        string `json:"name"`
	In          string `json:"in"`
}

type SwaggerInfo struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type SwaggerMethod struct {
	Tags []string `json:"tags,omitempty"`

	ProMethodName string                                   `json:"-"`
	Summary       string                                   `json:"summary"`
	Params        []SwaggerParameter                       `json:"parameters,omitempty"`
	RequestBody   map[string]map[string]SwaggerRequestBody `json:"requestBody,omitempty"`
	Responses     map[string]SwaggerResponsesDescription   `json:"responses"`
	Security      []map[string][]string                    `json:"security,omitempty"`
}

type SwaggerRequestBody struct {
	Schema SwaggerSchemaRef `json:"schema"`
}
type SwaggerSchemaRef struct {
	Ref        string                   `json:"$ref,omitempty"`
	Type       string                   `json:"type,omitempty"`
	Properties map[string]SwaggerSchema `json:"properties,omitempty"`
}

func (s *SwaggerRequestBody) GetClassName() string {
	cName := filepath.Base(s.Schema.Ref)
	return cName
}

type SwaggerParameter struct {
	Name        string        `json:"name"`
	In          string        `json:"in"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Schema      SwaggerSchema `json:"schema"`
}

type SwaggerSchema struct {
	Type  string `json:"type"`
	Items struct {
		Type   string `json:"type,omitempty"`
		Format string `json:"format,omitempty"`
	} `json:"items,omitempty"`
	Format   string      `json:"format,omitempty"`
	Default  interface{} `json:"default,omitempty"`
	Nullable bool        `json:"nullable,omitempty"`
}

type SwaggerResponsesDescription struct {
	Description string `json:"description"`
}

type OutClass struct {
	Class      string
	LowerClass string
	OutMethods []OutMethod
}
type OutMethod struct {
	Route     string
	Method    string
	ParamName string
	IsAuth    bool
}

// Build builds swagger json file  for given searchDir and mainAPIFile. Returns json
func (g *Gen) Build(config *Config) error {
	if _, err := os.Stat(config.SearchDir); os.IsNotExist(err) {
		return fmt.Errorf("dir: %s is not exist", config.SearchDir)
	}

	log.Println("Generate swagger docs....")
	p := NewParser(SetMarkdownFileDirectory(config.MarkdownFilesDir))
	p.PropNamingStrategy = config.PropNamingStrategy
	p.ParseVendor = config.ParseVendor
	p.ParseDependency = config.ParseDependency
	p.swagger.Info.Title = config.ProjectName
	p.swagger.Info.Version = "v1"
	if err := p.ParseAPI(config.SearchDir); err != nil {
		return err
	}
	swagger := p.GetSwagger()

	//b, err := g.jsonIndent(swagger)
	//if err != nil {
	//	return err
	//}

	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return err
	}

	docFileName := path.Join(config.OutputDir, "hiweb.go")
	//jsonFileName := path.Join(config.OutputDir, "swagger.json")
	//yamlFileName := path.Join(config.OutputDir, "swagger.yaml")

	docs, err := os.Create(docFileName)
	if err != nil {
		return err
	}
	defer docs.Close()

	// Write doc
	err = g.writeGoDoc(docs, swagger, config)
	if err != nil {
		return err
	}

	log.Printf("create docs.go at  %+v", docFileName)
	if config.VueBaseUrl != "" {
		apiDocFileName := config.VueOutputDir
		if config.VueOutputDir == "" {
			apiDocFileName = path.Join(config.OutputDir, "..", "..", "web", "src", "api", "api.js")
		}
		err = genVue(apiDocFileName, config.VueBaseUrl, swagger)
		if err != nil {
			return err
		}
		log.Printf("create api.js at  %+v", apiDocFileName)
	}

	//log.Printf("create swagger.json at  %+v", jsonFileName)
	//log.Printf("create swagger.yaml at  %+v", yamlFileName)

	return nil
}

func (g *Gen) writeFile(b []byte, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	return err
}

func (g *Gen) writeGoDoc(output io.Writer, swaggerSpec *SwaggerSpec, config *Config) error {
	generator, err := template.New("swagger_info").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			return v
			// Add schemes
			//v = "{\n    \"schemes\": {{ marshal .Schemes }}," + v[1:]
			// Sanitize backticks
			//return strings.Replace(v, "`", "`+\"`\"+`", -1)
		},
	}).Parse(packageTemplate)
	if err != nil {
		return err
	}
	packageName := filepath.Base(config.OutputDir)
	// crafted docs.json
	buf, err := g.jsonIndent(swaggerSpec)
	if err != nil {
		return err
	}

	outMethodMap := make(map[string]OutClass)
	for k, vs := range swaggerSpec.Paths {
		// tactions := strings.Split(k, "/")
		// actions := make([]string, 0)
		// for _, ta := range tactions {
		// 	if ta == "" {
		// 		continue
		// 	}
		// 	actions = append(actions, ta)
		// }
		// mName := strings.TrimSpace(actions[1])
		// cName := strings.TrimSpace(actions[0])
		// lcName := firstLower(cName)
		tsm, hasGet := vs["get"]
		_, hasPost := vs["post"]
		httpMethod := ""
		var sm SwaggerMethod
		if hasGet && hasPost {
			httpMethod = "*"
			sm = tsm
		} else {
			for tk, tv := range vs {
				httpMethod = tk
				sm = tv
				break
			}
		}
		paramNames := make([]string, 0)
		for _, p := range sm.Params {
			paramNames = append(paramNames, p.Name)
		}
		isAuth := false
		if len(sm.Security) > 0 {
			isAuth = true
		}
		cName := sm.Tags[0]
		lcName := firstLower(cName)
		route := k
		if i := strings.Index(route, "{"); i > 0 {
			route = route[:i]
		}
		// if len(actions) > 2 {
		// 	route = fmt.Sprintf("/%s/%s/", actions[0], actions[1])
		// } else {
		// 	route = fmt.Sprintf("/%s/%s", actions[0], actions[1])
		// }
		var has bool
		var outs OutClass
		if outs, has = outMethodMap[cName]; !has {
			outs = OutClass{Class: cName, LowerClass: lcName, OutMethods: make([]OutMethod, 0)}
		}
		outs.OutMethods = append(outs.OutMethods, OutMethod{
			Route:     route,
			Method:    httpMethod + ":" + sm.ProMethodName,
			ParamName: strings.Join(paramNames, ";"),
			IsAuth:    isAuth,
		})
		outMethodMap[cName] = outs
	}
	buffer := &bytes.Buffer{}
	err = generator.Execute(buffer, struct {
		PackageName   string
		ProjectName   string
		Timestamp     time.Time
		GeneratedTime bool
		Doc           string
		Methods       map[string]OutClass
	}{
		PackageName:   packageName,
		ProjectName:   config.ProjectName,
		Timestamp:     time.Now(),
		GeneratedTime: config.GeneratedTime,
		Doc:           string(buf),
		Methods:       outMethodMap,
	})
	if err != nil {
		return err
	}

	code := FormatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)
	return err
}

var packageTemplate = `// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag{{ if .GeneratedTime }} at
// {{ .Timestamp }}{{ end }}

package {{.PackageName}}

import (
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/autumnzw/hiweb"

	"github.com/alecthomas/template"
)

func init(){
	http.HandleFunc("/swag/", hiweb.Handler(
		hiweb.URL("./swagger.json","{{.ProjectName}}"), //The url pointing to API definition"
	))
}

var doc = ` + "`{{ printDoc .Doc}}`" + `

type swaggerInfo struct {
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{ 
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	hiweb.SwaggerRegister(&s{})
{{range $si,$vs := .Methods}}
	{{$vs.LowerClass}} := {{$vs.Class}}{}
{{range $i,$v := $vs.OutMethods}}
	hiweb.Route("{{$v.Route}}",&{{$vs.LowerClass}},"{{$v.ParamName}}","{{$v.Method}}",hiweb.RouteOption{IsAuth:{{$v.IsAuth}}})	
{{end}}	
{{end}}
}
`
