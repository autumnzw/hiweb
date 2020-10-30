package webcmd

import (
	"fmt"
	"go/ast"
	"go/build"
	goparser "go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Parser struct {
	// swagger represents the root document object for the API specification
	swagger *SwaggerSpec

	// files is a map that stores map[real_go_file_path][astFile]
	files map[string]*ast.File

	// TypeDefinitions is a map that stores [package name][type name][*ast.TypeSpec]
	TypeDefinitions map[string]map[string]*ast.TypeSpec

	// ImportAliases is map that stores [import name][import package name][*ast.ImportSpec]
	ImportAliases map[string]map[string]*ast.ImportSpec

	// CustomPrimitiveTypes is a map that stores custom primitive types to actual golang types [type name][string]
	CustomPrimitiveTypes map[string]string

	// registerTypes is a map that stores [refTypeName][*ast.TypeSpec]
	registerTypes map[string]*ast.TypeSpec

	PropNamingStrategy string

	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder
	ParseDependency bool

	// structStack stores full names of the structures that were already parsed or are being parsed now
	structStack []string

	// markdownFileDir holds the path to the folder, where markdown files are stored
	markdownFileDir string
}

// New creates a new Parser with default properties.
func NewParser(options ...func(*Parser)) *Parser {
	parser := &Parser{
		swagger: &SwaggerSpec{
			OpenApi: "",
			Info: SwaggerInfo{
				Title:   "",
				Version: "",
			},
			Paths: make(map[string]map[string]SwaggerMethod),
		},
		files:                make(map[string]*ast.File),
		TypeDefinitions:      make(map[string]map[string]*ast.TypeSpec),
		ImportAliases:        make(map[string]map[string]*ast.ImportSpec),
		CustomPrimitiveTypes: make(map[string]string),
		registerTypes:        make(map[string]*ast.TypeSpec),
	}

	for _, option := range options {
		option(parser)
	}

	return parser
}

// SetMarkdownFileDirectory sets the directory to search for markdownfiles
func SetMarkdownFileDirectory(directoryPath string) func(*Parser) {
	return func(p *Parser) {
		p.markdownFileDir = directoryPath
	}
}

func (parser *Parser) ParseAPI(searchDir string) error {
	Printf("Generate general API Info, search dir:%s", searchDir)

	parser.swagger.OpenApi = "3.0.1"

	if err := parser.getAllGoFileInfo(searchDir); err != nil {
		return err
	}

	for fileName, astFile := range parser.files {
		if err := parser.ParseRouterAPIInfo(fileName, astFile); err != nil {
			return err
		}
	}

	return nil //parser.parseDefinitions()
}

func getSwaggerSchemaField(f *ast.Field) SwaggerSchema {
	ss := SwaggerSchema{}
	switch t := f.Type.(type) {
	case *ast.ArrayType:
		ss.Type = "array"
		switch tt := t.Elt.(type) {
		case *ast.Ident:
			ss.Format = ""
			switch tt.Name {
			case "string":
				ss.Items.Type = "string"
			case "int":
				ss.Items.Type = "integer"
				ss.Items.Format = "int32"
			case "file":
				ss.Items.Type = "file"
			}
		}
	case *ast.Ident:
		ss.Format = ""
		switch t.Name {
		case "string":
			ss.Type = t.Name
		case "int":
			ss.Type = "integer"
			ss.Format = "int32"
		case "file":
			ss.Type = "file"
		}
	}
	return ss
}

func getSwaggerSchema(typeName string) SwaggerSchema {
	ss := SwaggerSchema{}
	ss.Format = ""
	switch typeName {
	case "string":
		ss.Type = typeName
	case "int":
		ss.Type = "integer"
		ss.Format = "int32"
	case "float64", "float32":
		ss.Type = "number"
		ss.Format = "float"
	case "file":
		ss.Type = "file"
	}
	return ss
}

// ParseRouterAPIInfo parses router api info for given astFile
func (parser *Parser) ParseRouterAPIInfo(fileName string, astFile *ast.File) error {
	for _, astDescription := range astFile.Decls {
		switch astDeclaration := astDescription.(type) {
		case *ast.FuncDecl:
			if astDeclaration.Recv != nil {
				hasAuth := false
				methodName := astDeclaration.Name.Name
				recvName := astDeclaration.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

				httpMethod := "get"
				route := ""
				sm := SwaggerMethod{
					Tags: []string{recvName},
					Responses: map[string]SwaggerResponsesDescription{
						"200": {"Success"},
						"401": {"Unauthorized"},
						"403": {"Forbidden"},
					},
					ProMethodName: methodName,
					Security:      []map[string][]string{},
				}
				var cm *SwaggerComponent
				if parser.swagger.Components != nil {
					cm = parser.swagger.Components
				} else {
					cm = &SwaggerComponent{
						Schema: map[string]SwaggerComponentStruct{},
					}
				}
				paramMap := make(map[string]SwaggerParameter)
				if astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
					operation := NewOperation() //for per 'function' comment, create a new 'Operation' object
					operation.parser = parser
					for _, comment := range astDeclaration.Doc.List {
						if err := operation.ParseComment(comment.Text, astFile); err != nil {
							return fmt.Errorf("ParseComment error in file %s :%+v", fileName, err)
						}
					}
					sm.Summary = operation.Summary
					sm.Security = operation.Security
					if len(sm.Security) > 0 {
						hasAuth = true
					}
					for _, p := range operation.Params {
						paramMap[p.Name] = p
					}
					httpMethod = operation.HTTPMethod
					route = operation.Path
				}
				urlParam := ""
				sm.Params = make([]SwaggerParameter, 0)
				sm.RequestBody = make(map[string]map[string]SwaggerRequestBody)
				paramLen := 0
				//添加上传注释
				for _, v := range paramMap {
					if v.In == "formData" {
						sm.RequestBody["content"] = make(map[string]SwaggerRequestBody)
						sm.RequestBody["content"]["multipart/form-data"] = SwaggerRequestBody{Schema: SwaggerSchemaRef{
							Type:       "object",
							Properties: map[string]SwaggerSchema{v.Name: {Type: "string", Format: "binary"}},
						}}
					}
				}
				//参数中添加
				for _, param := range astDeclaration.Type.Params.List {
					paramLen += len(param.Names)
				}
				for _, param := range astDeclaration.Type.Params.List {
					for _, paramName := range param.Names {
						name := paramName.Name
						var typeObj *ast.Ident
						switch tTypeObj := param.Type.(type) {
						case *ast.Ident:
							typeObj = tTypeObj
						case *ast.StarExpr:
							typeObj = tTypeObj.X.(*ast.Ident)
						default:
							panic("type not support ast type")
						}
						paramTypeName := typeObj.Name
						if typeObj.Obj == nil {
							ss := getSwaggerSchema(paramTypeName)
							sp := paramMap[name]
							in := "query"
							if paramLen == 1 && httpMethod == "get" && name == "key" {
								urlParam = "{" + name + "}"
								in = "path"
							} else {
								in = "query"
							}

							sm.Params = append(sm.Params, SwaggerParameter{
								Name:        name,
								Schema:      ss,
								In:          in,
								Description: sp.Description,
							})
						} else {
							refObjName := fmt.Sprintf("#/components/schemas/%s", paramTypeName)
							sm.RequestBody["content"] = make(map[string]SwaggerRequestBody)
							sm.RequestBody["content"]["application/json-patch+json"] = SwaggerRequestBody{Schema: SwaggerSchemaRef{Ref: refObjName}}
							sm.RequestBody["content"]["application/json"] = SwaggerRequestBody{Schema: SwaggerSchemaRef{Ref: refObjName}}
							sm.RequestBody["content"]["text/json"] = SwaggerRequestBody{Schema: SwaggerSchemaRef{Ref: refObjName}}
							sm.RequestBody["content"]["application/*+json"] = SwaggerRequestBody{Schema: SwaggerSchemaRef{Ref: refObjName}}

							prop := map[string]SwaggerSchema{}
							for _, f := range typeObj.Obj.Decl.(*ast.TypeSpec).Type.(*ast.StructType).Fields.List {
								propName := getTagName(f.Tag.Value)
								if propName == "" {
									propName = f.Names[0].Name
								}
								prop[propName] = getSwaggerSchemaField(f)
							}
							cm.Schema[paramTypeName] = SwaggerComponentStruct{
								Type:       "object",
								Properties: prop,
							}
						}
					}

				}
				if route == "" {
					if len(urlParam) > 0 {
						route = fmt.Sprintf("/%s/%s/%s", recvName, methodName, urlParam)
					} else {
						route = fmt.Sprintf("/%s/%s", recvName, methodName)
					}
				}

				if parser.swagger.Paths[route] == nil {
					parser.swagger.Paths[route] = map[string]SwaggerMethod{}
				} else {
					return fmt.Errorf("err same route file:%s", fileName)
				}
				if hasAuth {
					cm.SecuritySchemes = map[string]SwaggerComponentSecuritySchemes{
						"oauth2": {
							Type:        "apiKey",
							Description: "JWT授权(数据将在请求头中进行传输) 直接在下框中输入Bearer {token}（注意两者之间是一个空格）\"",
							Name:        "Authorization",
							In:          "header",
						},
					}
				}
				if httpMethod == "getpost" {
					parser.swagger.Paths[route]["get"] = sm
					parser.swagger.Paths[route]["post"] = sm
				} else {
					parser.swagger.Paths[route][httpMethod] = sm
				}

				parser.swagger.Components = cm
			}

		}
	}

	return nil
}

// GetAllGoFileInfo gets all Go source files information for given searchDir.
func (parser *Parser) getAllGoFileInfo(searchDir string) error {
	return filepath.Walk(searchDir, parser.visit)
}

func (parser *Parser) visit(path string, f os.FileInfo, err error) error {
	if err := parser.Skip(path, f); err != nil {
		return err
	}
	if f.Name() == "hiweb.go" {
		return nil
	} else {
		return parser.parseFile(path)
	}
}

// Skip returns filepath.SkipDir error if match vendor and hidden folder
func (parser *Parser) Skip(path string, f os.FileInfo) error {

	if !parser.ParseVendor { // ignore vendor
		if f.IsDir() && f.Name() == "vendor" {
			return filepath.SkipDir
		}
	}

	// exclude all hidden folder
	if f.IsDir() && len(f.Name()) > 1 && f.Name()[0] == '.' {
		return filepath.SkipDir
	}
	return nil
}

func (parser *Parser) parseFile(path string) error {
	if ext := filepath.Ext(path); ext == ".go" {
		fset := token.NewFileSet() // positions are relative to fset
		astFile, err := goparser.ParseFile(fset, path, nil, goparser.ParseComments)
		if err != nil {
			return fmt.Errorf("ParseFile error:%+v", err)
		}

		parser.files[path] = astFile
	}
	return nil
}

func getPkgName(searchDir string) (string, error) {
	cmd := exec.Command("go", "list", "-f={{.ImportPath}}")
	cmd.Dir = searchDir
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("execute go list command, %s, stdout:%s, stderr:%s", err, stdout.String(), stderr.String())
	}

	outStr, _ := stdout.String(), stderr.String()

	if outStr[0] == '_' { // will shown like _/{GOPATH}/src/{YOUR_PACKAGE} when NOT enable GO MODULE.
		outStr = strings.TrimPrefix(outStr, "_"+build.Default.GOPATH+"/src/")
	}
	f := strings.Split(outStr, "\n")
	outStr = f[0]

	return outStr, nil
}

func (parser *Parser) isInStructStack(refTypeName string) bool {
	for _, structName := range parser.structStack {
		if refTypeName == structName {
			return true
		}
	}
	return false
}

func fullTypeName(pkgName, typeName string) string {
	if pkgName != "" {
		return pkgName + "." + typeName
	}
	return typeName
}

// GetSwagger returns *spec.Swagger which is the root document object for the API specification.
func (parser *Parser) GetSwagger() *SwaggerSpec {
	return parser.swagger
}
