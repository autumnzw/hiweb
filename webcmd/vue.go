package webcmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"
)

var vueTemplate = `
import BAPI from './bapi'

{{range $i,$v := .Methods}}
function {{$v.MethodName}}({{$v.ParamNames}}){

	let tmpUrl = "{{$v.MethodPath}}";

	{{if eq  $v.MethodType "post"}}
		let inparam={
		{{range $is,$vs := $v.ParamList}}
			"{{$vs}}":{{$vs}},
		{{end}}
		}
	{{else}}
		{{range $is,$vs := $v.ParamList}}
		tmpUrl = BAPI.AppendParam(tmpUrl, '{{$vs}}', {{$vs}}) 
		{{end}}	
	{{end}}	
	
	return BAPI.Xhr({
		url: tmpUrl,
		method: '{{$v.MethodType}}',
	{{if eq  $v.MethodType "post"}}
		body:inparam,
	{{end}}
	}).then((data) => {
		return data
	})

}
{{end}}	

{{range $i,$v := .Methods}}
export{ {{$v.MethodName}} }
{{end}}	
`

type VueFunction struct {
	MethodName string
	MethodPath string
	ParamNames string
	ParamList  []string
	MethodType string

	IsAuth bool
}

func genVue(apiDocFileName string, vueBaseUrl string, swaggerSpec *SwaggerSpec) error {
	apiDoc, err := os.Create(apiDocFileName)
	if err != nil {
		return err
	}
	defer apiDoc.Close()

	generator, err := template.New("swagger_vue_info").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			return v
			// Add schemes
			//v = "{\n    \"schemes\": {{ marshal .Schemes }}," + v[1:]
			// Sanitize backticks
			//return strings.Replace(v, "`", "`+\"`\"+`", -1)
		},
	}).Parse(vueTemplate)
	if err != nil {
		return err
	}

	outMethodList := make([]VueFunction, 0)
	for k, vs := range swaggerSpec.Paths {
		tactions := strings.Split(k, "/")
		actions := make([]string, 0)
		for _, ta := range tactions {
			if ta == "" {
				continue
			}
			actions = append(actions, ta)
		}
		for tk, tv := range vs {
			inClassName := ""
			for _, rbv := range tv.RequestBody {
				for _, srbv := range rbv {
					inClassName = srbv.GetClassName()
					break
				}
			}
			paramNames := make([]string, 0)
			for _, p := range tv.Params {
				paramNames = append(paramNames, p.Name)
			}
			if inClassName != "" {
				sClass := swaggerSpec.Components.Schema[inClassName]
				for pk := range sClass.Properties {
					paramNames = append(paramNames, pk)
				}
			}
			isAuth := false
			if len(tv.Security) > 0 {
				isAuth = true
			}
			methodName := fmt.Sprintf("%s%s", actions[0], actions[1])
			methodPath := fmt.Sprintf("/%s/%s", actions[0], actions[1])
			outMethodList = append(outMethodList, VueFunction{
				MethodName: methodName,
				MethodPath: methodPath,
				MethodType: tk,
				ParamNames: strings.Join(paramNames, ","),
				ParamList:  paramNames,
				IsAuth:     isAuth,
			})

		}
	}

	buffer := &bytes.Buffer{}
	err = generator.Execute(buffer, struct {
		BaseUrl   string
		Timestamp time.Time
		Methods   []VueFunction
	}{
		BaseUrl:   vueBaseUrl,
		Timestamp: time.Now(),
		Methods:   outMethodList,
	})
	if err != nil {
		return err
	}

	code := FormatSource(buffer.Bytes())

	// write
	_, err = apiDoc.Write(code)
	return err
}
