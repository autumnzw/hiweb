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
import axios from 'axios'
//切换路由实现进度条
import NProgress from 'nprogress'
// import 'nprogress/nprogress.css'

let http = axios.create({
  withCredentials: false,
  headers: {
      'Content-Type': 'application/json;charset=UTF-8'
  },
});

http.defaults.withCredentials = false;//跨域安全策略

// 请求拦截器
http.interceptors.request.use(
  config => {
    let token = GetLocalToken();
    if (token != "" && config.url.indexOf("/Token?") == -1) {
      config.headers.Authorization = token;
      NProgress.start() 
    }
  
    return config;
  },
  err => {
      return Promise.reject(err);
  }
);

let baseUrl="{{.BaseUrl}}"
function InitUrl(url){
  if(url.endWidth("/")){
    baseUrl=url
  }else{
    baseUrl=url+"/"
  }
  console.log('url:',baseUrl)
}
function Xhr( {url, body = null, method = 'get'} ) {
  if(method == "get"){
    return http
      .get(baseUrl+url)
      .then(response => response)
      .catch(handleError);
  }else if(method=="post"){
    return http
      .post(baseUrl+url,body)
      .then(response => response)
      .catch(handleError);
  }else{
    throw new TypeError("not support method", method)
  }
}

var isLogin = false;
function SetLocalToken(token){
  let tokenStr = JSON.stringify(token);
  window.localStorage.setItem('Token', tokenStr)
  isLogin = true;
}

function IsLogin(){
  return isLogin;
}

function GetLocalToken(){
  //   取出刚才存储在本地的sid值赋值给localsid
  const token = window.localStorage.getItem('Token')
  if(token) {
    try{
      const value=JSON.parse(token)
      const expires = new Date(value.profile.expires_at * 1000)
      const now = new Date().getTime();
      const isExpire = now > expires.getTime();
      if (isExpire){
          return ""
      }else{
          return value.token_type + " " + value.access_token;
      }
    }catch{
      return "";
    }
    
  }
  return ""
}

function GetUserName(){
	const token = window.localStorage.getItem('Token')
	if(token) {
		try{
			const value=JSON.parse(token)
			return value.user_name
		}catch{
			return "";
		}
	}
	return ""
}

function AppendParam(url, name, value) {
  if (url && name) {
      name += '=';
      if (url.indexOf(name) === -1) {
          if (url.indexOf('?') !== -1) {
              url += '&';
          } else {
              url += '?';
          }
          url += name + encodeURIComponent(value);
      }
  }
  return url;
}
export {SetLocalToken,GetLocalToken,GetUserName,IsLogin,AppendParam,Xhr,InitUrl}

function handleError(error) {
  let errMsg;
  if (error instanceof Response) {
      if (error.status === 405) {
          return Promise.reject(errMsg);
      }
      const body = error.json() || '';
      const err = body.error || JSON.stringify(body);
      errMsg = "${error.status} - ${error.statusText || ''} "+err;
  } else {
      errMsg = error.message ? error.message : error.toString();
  }
  // console.error(errMsg);
  return Promise.reject(errMsg);
}

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
		tmpUrl = AppendParam(tmpUrl, '{{$vs}}', {{$vs}}) 
		{{end}}	
	{{end}}	
	
	return Xhr({
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
