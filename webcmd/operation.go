package webcmd

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"golang.org/x/tools/go/loader"
)

// Operation describes a single API operation on a path.
// For more information: https://github.com/swaggo/swag#api-operation
type Operation struct {
	HTTPMethod string
	Path       string
	SwaggerMethod

	parser *Parser
}

var mimeTypeAliases = map[string]string{
	"json":                  "application/json",
	"xml":                   "text/xml",
	"plain":                 "text/plain",
	"html":                  "text/html",
	"mpfd":                  "multipart/form-data",
	"x-www-form-urlencoded": "application/x-www-form-urlencoded",
	"json-api":              "application/vnd.api+json",
	"json-stream":           "application/x-json-stream",
	"octet-stream":          "application/octet-stream",
	"png":                   "image/png",
	"jpeg":                  "image/jpeg",
	"gif":                   "image/gif",
}

var mimeTypePattern = regexp.MustCompile("^[^/]+/[^/]+$")

// NewOperation creates a new Operation with default properties.
// map[int]Response
func NewOperation() *Operation {
	return &Operation{
		HTTPMethod: "get",
		SwaggerMethod: SwaggerMethod{
			Params:      []SwaggerParameter{},
			RequestBody: map[string]map[string]SwaggerRequestBody{},
			Security:    []map[string][]string{},
		},
	}
}

// ParseComment parses comment for given comment string and returns error if error occurs.
func (operation *Operation) ParseComment(comment string, astFile *ast.File) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}
	attribute := strings.Fields(commentLine)[0]
	lineRemainder := strings.TrimSpace(commentLine[len(attribute):])
	lowerAttribute := strings.ToLower(attribute)

	var err error
	switch lowerAttribute {
	case "@description":
		operation.ParseDescriptionComment(lineRemainder)
	// case "@router":
	// 	err = operation.ParseRouterComment(lineRemainder)
	case "@param":
		err = operation.ParseParamComment(lineRemainder, "query", astFile)
	case "@auth":
		err = operation.ParseAuthComment(lineRemainder)
	case "@httpget":
		err = operation.ParseHttpGetComment(lineRemainder)
	case "@httppost":
		err = operation.ParseHttpPostComment(lineRemainder)
	case "@httpgetpost":
		err = operation.ParseHttpGetPostComment(lineRemainder)
	case "@httpdelete":
		err = operation.ParseHttpDeleteComment(lineRemainder)
	case "@httpput":
		err = operation.ParseHttpPutComment(lineRemainder)
	case "@upload":
		err = operation.ParseParamComment(lineRemainder, "formData", astFile)
	default:
		err = operation.ParseMetadata(attribute, lowerAttribute, lineRemainder)
	}

	return err
}

// ParseDescriptionComment godoc
func (operation *Operation) ParseDescriptionComment(lineRemainder string) {
	if operation.Summary == "" {
		operation.Summary = lineRemainder
		return
	}
	operation.Summary += "\n" + lineRemainder
}

// ParseMetadata godoc
func (operation *Operation) ParseMetadata(attribute, lowerAttribute, lineRemainder string) error {
	// parsing specific meta data extensions
	if strings.HasPrefix(lowerAttribute, "@x-") {
		if len(lineRemainder) == 0 {
			return fmt.Errorf("annotation %s need a value", attribute)
		}

		var valueJSON interface{}
		if err := json.Unmarshal([]byte(lineRemainder), &valueJSON); err != nil {
			return fmt.Errorf("annotation %s need a valid json value", attribute)
		}
		//operation.Operation.AddExtension(attribute[1:], valueJSON) // Trim "@" at head
	}
	return nil
}

var paramPattern = regexp.MustCompile(`(\S+)[\s]+([\w]+)[\s]+([\S.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

// ParseParamComment parses params return []string of param properties
// E.g. @Param	queryText		formData	      string	  true		        "The email for login"
//              [param name]    [paramType] [data type]  [is mandatory?]   [Comment]
// E.g. @Param   some_id     path    int     true        "Some ID"
func (operation *Operation) ParseParamComment(commentLine string, inType string, astFile *ast.File) error {
	matches := strings.SplitN(commentLine, " ", 2)
	if len(matches) < 2 {
		return fmt.Errorf("param len is min 2:%s", commentLine)
	}
	operation.Params = append(operation.Params, SwaggerParameter{
		Name:        matches[0],
		Description: matches[1],
		In:          inType,
	})
	return nil
}

func (operation *Operation) registerSchemaType(schemaType string, astFile *ast.File) (string, *ast.TypeSpec, error) {
	if !strings.ContainsRune(schemaType, '.') {
		if astFile == nil {
			return schemaType, nil, fmt.Errorf("no package name for type %s", schemaType)
		}
		schemaType = astFile.Name.String() + "." + schemaType
	}
	refSplit := strings.Split(schemaType, ".")
	pkgName := refSplit[0]
	typeName := refSplit[1]
	if typeSpec, ok := operation.parser.TypeDefinitions[pkgName][typeName]; ok {
		operation.parser.registerTypes[schemaType] = typeSpec
		return schemaType, typeSpec, nil
	}
	var typeSpec *ast.TypeSpec
	if astFile == nil {
		return schemaType, nil, fmt.Errorf("can not register schema type: %q reason: astFile == nil", schemaType)
	}
	for _, imp := range astFile.Imports {
		if imp.Name != nil && imp.Name.Name == pkgName { // the import had an alias that matched
			break
		}
		impPath := strings.Replace(imp.Path.Value, `"`, ``, -1)
		if strings.HasSuffix(impPath, "/"+pkgName) {
			var err error
			typeSpec, err = findTypeDef(impPath, typeName)
			if err != nil {
				return schemaType, nil, fmt.Errorf("can not find type def: %q error: %s", schemaType, err)
			}
			break
		}
	}

	if typeSpec == nil {
		return schemaType, nil, fmt.Errorf("can not find schema type: %q", schemaType)
	}

	if _, ok := operation.parser.TypeDefinitions[pkgName]; !ok {
		operation.parser.TypeDefinitions[pkgName] = make(map[string]*ast.TypeSpec)
	}

	operation.parser.TypeDefinitions[pkgName][typeName] = typeSpec
	operation.parser.registerTypes[schemaType] = typeSpec
	return schemaType, typeSpec, nil
}

var regexAttributes = map[string]*regexp.Regexp{
	// for Enums(A, B)
	"enums": regexp.MustCompile(`(?i)enums\(.*\)`),
	// for Minimum(0)
	"maxinum": regexp.MustCompile(`(?i)maxinum\(.*\)`),
	// for Maximum(0)
	"mininum": regexp.MustCompile(`(?i)mininum\(.*\)`),
	// for Maximum(0)
	"default": regexp.MustCompile(`(?i)default\(.*\)`),
	// for minlength(0)
	"minlength": regexp.MustCompile(`(?i)minlength\(.*\)`),
	// for maxlength(0)
	"maxlength": regexp.MustCompile(`(?i)maxlength\(.*\)`),
	// for format(email)
	"format": regexp.MustCompile(`(?i)format\(.*\)`),
}

func (operation *Operation) parseAndExtractionParamAttribute(commentLine, schemaType string, param *spec.Parameter) error {
	schemaType = TransToValidSchemeType(schemaType)
	for attrKey, re := range regexAttributes {
		attr, err := findAttr(re, commentLine)
		if err != nil {
			continue
		}
		switch attrKey {
		case "enums":
			err := setEnumParam(attr, schemaType, param)
			if err != nil {
				return err
			}
		case "maxinum":
			n, err := setNumberParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.Maximum = &n
		case "mininum":
			n, err := setNumberParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.Minimum = &n
		case "default":
			value, err := defineType(schemaType, attr)
			if err != nil {
				return nil
			}
			param.Default = value
		case "maxlength":
			n, err := setStringParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.MaxLength = &n
		case "minlength":
			n, err := setStringParam(attrKey, schemaType, attr, commentLine)
			if err != nil {
				return err
			}
			param.MinLength = &n
		case "format":
			param.Format = attr
		}

	}
	return nil
}

func findAttr(re *regexp.Regexp, commentLine string) (string, error) {
	attr := re.FindString(commentLine)
	l := strings.Index(attr, "(")
	r := strings.Index(attr, ")")
	if l == -1 || r == -1 {
		return "", fmt.Errorf("can not find regex=%s, comment=%s", re.String(), commentLine)
	}
	return strings.TrimSpace(attr[l+1 : r]), nil
}

func setStringParam(name, schemaType, attr, commentLine string) (int64, error) {
	if schemaType != "string" {
		return 0, fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
	n, err := strconv.ParseInt(attr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s is allow only a number got=%s", name, attr)
	}
	return n, nil
}

func setNumberParam(name, schemaType, attr, commentLine string) (float64, error) {
	if schemaType != "integer" && schemaType != "number" {
		return 0, fmt.Errorf("%s is attribute to set to a number. comment=%s got=%s", name, commentLine, schemaType)
	}
	n, err := strconv.ParseFloat(attr, 64)
	if err != nil {
		return 0, fmt.Errorf("maximum is allow only a number. comment=%s got=%s", commentLine, attr)
	}
	return n, nil
}

func setEnumParam(attr, schemaType string, param *spec.Parameter) error {
	for _, e := range strings.Split(attr, ",") {
		e = strings.TrimSpace(e)

		value, err := defineType(schemaType, e)
		if err != nil {
			return err
		}
		param.Enum = append(param.Enum, value)
	}
	return nil
}

// defineType enum value define the type (object and array unsupported)
func defineType(schemaType string, value string) (interface{}, error) {
	schemaType = TransToValidSchemeType(schemaType)
	switch schemaType {
	case "string":
		return value, nil
	case "number":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	case "integer":
		v, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	case "boolean":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("enum value %s can't convert to %s err: %s", value, schemaType, err)
		}
		return v, nil
	default:
		return nil, fmt.Errorf("%s is unsupported type in enum value", schemaType)
	}
}

// ParseTagsComment parses comment for given `tag` comment string.
func (operation *Operation) ParseTagsComment(commentLine string) {
	//tags := strings.Split(commentLine, ",")
	//for _, tag := range tags {
	//operation.SwaggerParameter. = append(operation.Tags, strings.TrimSpace(tag))
	//}
}

// ParseAcceptComment parses comment for given `accept` comment string.
//func (operation *Operation) ParseAcceptComment(commentLine string) error {
//	return parseMimeTypeList(commentLine, &operation.Consumes, "%v accept type can't be accepted")
//}

// ParseProduceComment parses comment for given `produce` comment string.
//func (operation *Operation) ParseProduceComment(commentLine string) error {
//	return parseMimeTypeList(commentLine, &operation.Produces, "%v produce type can't be accepted")
//}

// parseMimeTypeList parses a list of MIME Types for a comment like
// `produce` (`Content-Type:` response header) or
// `accept` (`Accept:` request header)
func parseMimeTypeList(mimeTypeList string, typeList *[]string, format string) error {
	mimeTypes := strings.Split(mimeTypeList, ",")
	for _, typeName := range mimeTypes {
		if mimeTypePattern.MatchString(typeName) {
			*typeList = append(*typeList, typeName)
			continue
		}
		if aliasMimeType, ok := mimeTypeAliases[typeName]; ok {
			*typeList = append(*typeList, aliasMimeType)
			continue
		}
		return fmt.Errorf(format, typeName)
	}
	return nil
}

var routerPattern = regexp.MustCompile(`^(/[\w\.\/\-{}\+:]*)[[:blank:]]+\[(\w+)]`)

// ParseRouterComment parses comment for gived `router` comment string.
func (operation *Operation) ParseRouterComment(commentLine string) error {
	var matches []string

	if matches = routerPattern.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("can not parse router comment \"%s\"", commentLine)
	}
	path := matches[1]
	httpMethod := matches[2]

	operation.Path = path
	operation.HTTPMethod = strings.ToUpper(httpMethod)

	return nil
}

// ParseAuthComment parses comment for gived `security` comment string.
func (operation *Operation) ParseAuthComment(commentLine string) error {
	m := map[string][]string{}
	m["oauth2"] = []string{}
	operation.Security = append(operation.Security, m)
	return nil
}

func (operation *Operation) ParseHttpGetComment(commentLine string) error {
	operation.HTTPMethod = "get"
	if strings.HasPrefix(commentLine, "/") {
		operation.Path = commentLine
	}
	return nil
}

func (operation *Operation) ParseHttpPostComment(commentLine string) error {
	operation.HTTPMethod = "post"
	if strings.HasPrefix(commentLine, "/") {
		operation.Path = commentLine
	}
	return nil
}

func (operation *Operation) ParseHttpGetPostComment(commentLine string) error {
	operation.HTTPMethod = "getpost"
	if strings.HasPrefix(commentLine, "/") {
		operation.Path = commentLine
	}
	return nil
}

func (operation *Operation) ParseHttpDeleteComment(commentLine string) error {
	operation.HTTPMethod = "delete"
	if strings.HasPrefix(commentLine, "/") {
		operation.Path = commentLine
	}
	return nil
}

func (operation *Operation) ParseHttpPutComment(commentLine string) error {
	operation.HTTPMethod = "put"
	if strings.HasPrefix(commentLine, "/") {
		operation.Path = commentLine
	}
	return nil
}

// findTypeDef attempts to find the *ast.TypeSpec for a specific type given the
// type's name and the package's import path
// TODO: improve finding external pkg
func findTypeDef(importPath, typeName string) (*ast.TypeSpec, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	conf := loader.Config{
		ParserMode: goparser.SpuriousErrors,
		Cwd:        cwd,
	}

	conf.Import(importPath)

	lprog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	// If the pkg is vendored, the actual pkg path is going to resemble
	// something like "{importPath}/vendor/{importPath}"
	for k := range lprog.AllPackages {
		realPkgPath := k.Path()

		if strings.Contains(realPkgPath, "vendor/"+importPath) {
			importPath = realPkgPath
		}
	}

	pkgInfo := lprog.Package(importPath)

	if pkgInfo == nil {
		return nil, fmt.Errorf("package was nil")
	}

	// TODO: possibly cache pkgInfo since it's an expensive operation

	for i := range pkgInfo.Files {
		for _, astDeclaration := range pkgInfo.Files[i].Decls {
			if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
				for _, astSpec := range generalDeclaration.Specs {
					if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == typeName {
							return typeSpec, nil
						}
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("type spec not found")
}

var responsePattern = regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([\w\-\.\/]+)[^"]*(.*)?`)

var emptyResponsePattern = regexp.MustCompile(`([\d]+)[\s]+"(.*)"`)

// createParameter returns swagger spec.Parameter for gived  paramType, description, paramName, schemaType, required
func createParameter(paramType, description, paramName, schemaType string, required bool) spec.Parameter {
	// //five possible parameter types. 	query, path, body, header, form
	paramProps := spec.ParamProps{
		Name:        paramName,
		Description: description,
		Required:    required,
		In:          paramType,
	}
	if paramType == "body" {
		paramProps.Schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{schemaType},
			},
		}
		parameter := spec.Parameter{
			ParamProps: paramProps,
		}
		return parameter
	}
	parameter := spec.Parameter{
		ParamProps: paramProps,
		SimpleSchema: spec.SimpleSchema{
			Type: schemaType,
		},
	}
	return parameter
}
