package tinyweb

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type RouteOption struct {
	IsAuth bool
}

func Route(rootpath string, obj ControllerInterface, paramNames string, mappingMethod string, option RouteOption) {
	t := reflect.TypeOf(obj)
	params := strings.Split(paramNames, ";")
	fms := strings.Split(mappingMethod, ":")
	//httpMethod := fms[0]
	funcMethod := fms[1]
	isUrlParam := false
	if strings.HasSuffix(rootpath, "/") {
		isUrlParam = true
	}
	http.HandleFunc(rootpath, func(writer http.ResponseWriter, req *http.Request) {
		headers := writer.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		headers.Set("Access-Control-Allow-Headers", "*")
		headers.Set("Access-Control-Allow-Method", "*")
		headers.Set("Access-Control-Allow-Credentials", "true")
		if strings.ToLower(req.Method) == "options" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		if strings.ToLower(req.Method) != strings.ToLower(fms[0]) || fms[0] == "*" {
			writer.WriteHeader(http.StatusNotFound)
			fmt.Fprint(writer, "not found")
			return
		}
		if option.IsAuth {
			token, err := request.ParseFromRequest(req, request.AuthorizationHeaderExtractor,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(WebConfig.SecretKey), nil
				})
			if err == nil {
				if token.Valid {

				} else {
					writer.WriteHeader(http.StatusUnauthorized)
					fmt.Fprint(writer, "Token is not valid")
					return
				}
			} else {
				writer.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(writer, "Unauthorized access to this resource")
				return
			}
		}
		vc := reflect.New(t.Elem())
		execController, ok := vc.Interface().(ControllerInterface)
		if !ok {
			panic("controller is not ControllerInterface")
		}
		context := WebContext{req, writer}
		execController.Init(&context)
		m := vc.MethodByName(funcMethod)
		paramLen := m.Type().NumIn()
		var parameters []reflect.Value
		var err error
		if isUrlParam {
			tactions := strings.Split(req.RequestURI, "/")
			paramIn := make([]string, 0)
			for i, ta := range tactions {
				if ta == "" {
					continue
				}
				if i > 2 {
					paramIn = append(paramIn, ta)
				}
			}
			parameters, err = genParameters(m, params, paramLen, execController, paramIn)
		} else {
			parameters, err = genParameters(m, params, paramLen, execController, []string{})
		}
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(writer, "参数错误")
			WebConfig.Logger.Error("url:%s param err:%s", req.RequestURI, err)
			return
		}
		m.Call(parameters)
	})
}

func genParameters(m reflect.Value, params []string, paramLen int, execController ControllerInterface, paramIn []string) ([]reflect.Value, error) {
	parameters := make([]reflect.Value, 0, paramLen)
	for i := 0; i < paramLen; i++ {
		arg := m.Type().In(i)
		param := params[i]
		var paramVal interface{} = nil
		var err error
		if len(paramIn) > i {
			paramVal = paramIn[i]
		}
		//fmt.Printf("argument %d is %s[%s] type \n", i, arg.Name(), param)
		switch arg.Kind() {
		case reflect.Int:
			if paramVal == nil {
				paramVal, err = execController.Query(param)
				if err != nil {
					return parameters, fmt.Errorf("query err:%s", err)
				}
			}
			switch paramVal.(type) {
			case string:
				if paramVal == "" {
					parameters = append(parameters, reflect.ValueOf(0))
				} else {
					v, err := strconv.Atoi(paramVal.(string)) //strconv.ParseInt(paramVal.(string), 10, 32)
					if err != nil {
						return parameters, fmt.Errorf("argument %d %s convert int failed, %v \n", i, paramVal, err)
					}
					parameters = append(parameters, reflect.ValueOf(v))
				}
			case int:
				parameters = append(parameters, reflect.ValueOf(paramVal))
			case int64:
				parameters = append(parameters, reflect.ValueOf(int(paramVal.(int64))))
			case nil:
				parameters = append(parameters, reflect.ValueOf(0))
			}

		case reflect.String:
			if paramVal == nil {
				paramVal, err = execController.Query(param)
				if err != nil {
					return parameters, fmt.Errorf("query err:%s", err)
				}
			}
			switch paramVal.(type) {
			case string:
				parameters = append(parameters, reflect.ValueOf(paramVal))
			case int:
				v := strconv.Itoa(paramVal.(int))
				parameters = append(parameters, reflect.ValueOf(v))
			case nil:
				parameters = append(parameters, reflect.ValueOf(""))
			}
		case reflect.Struct:
			argObj := reflect.New(arg)
			err := execController.ParseValid(argObj.Interface())
			if err != nil {
				return parameters, fmt.Errorf("parse err:%s", err)
			}
			parameters = append(parameters, argObj.Elem())
		case reflect.Ptr:
			argObj := reflect.New(arg.Elem())
			err := execController.ParseValid(argObj.Interface())
			if err != nil {
				return parameters, fmt.Errorf("parse err:%s", err)
			}
			parameters = append(parameters, argObj)
		default:
			WebConfig.Logger.Error("unsupport type %s[%s] \n", arg.Kind(), param)
			return parameters, fmt.Errorf("unsupport type %s[%s] \n", arg.Kind(), param)
		}
	}
	return parameters, nil
}

func RouteFiles(route, dir string) {
	handler := http.FileServer(http.Dir(dir))
	http.Handle(route, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		WebConfig.Logger.Info("Started %s %s ip:%s", r.Method, r.URL.Path, r.RemoteAddr)
		handler.ServeHTTP(w, r)
		WebConfig.Logger.Info("Comleted %s in %v", r.URL.Path, time.Since(start))
	}))
}
func Map(obj ControllerInterface) error {
	methodNames := make([]string, 0)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	structName := t.Elem().Name()
	methodNum := v.NumMethod()
	for i := 0; i < methodNum; i++ {
		m := t.Method(i)
		if m.Index == 1 {
			methodNames = append(methodNames, m.Name)
		}
	}

	for _, methodName := range methodNames {
		routeName := fmt.Sprintf("/%s/%s", structName, methodName)
		http.HandleFunc(routeName, func(writer http.ResponseWriter, request *http.Request) {
			vc := reflect.New(t.Elem())
			execController, ok := vc.Interface().(ControllerInterface)
			if !ok {
				panic("controller is not ControllerInterface")
			}
			context := WebContext{request, writer}
			execController.Init(&context)
			vc.MethodByName(methodName).Call([]reflect.Value{})
		})
	}
	return nil
}

func JwtMap(obj ControllerInterface) error {
	methodNames := make([]string, 0)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	structName := t.Elem().Name()
	methodNum := v.NumMethod()
	for i := 0; i < methodNum; i++ {
		m := t.Method(i)
		if m.Index == 1 {
			methodNames = append(methodNames, m.Name)
		}
	}

	for _, methodName := range methodNames {
		routeName := fmt.Sprintf("/%s/%s", structName, methodName)
		http.HandleFunc(routeName, func(writer http.ResponseWriter, req *http.Request) {
			token, err := request.ParseFromRequest(req, request.AuthorizationHeaderExtractor,
				func(token *jwt.Token) (interface{}, error) {
					return []byte(WebConfig.SecretKey), nil
				})
			if err == nil {
				if token.Valid {
					vc := reflect.New(t.Elem())
					execController, ok := vc.Interface().(ControllerInterface)
					if !ok {
						panic("controller is not ControllerInterface")
					}
					context := WebContext{req, writer}
					execController.Init(&context)
					vc.MethodByName(methodName).Call([]reflect.Value{})
				} else {
					writer.WriteHeader(http.StatusUnauthorized)
					fmt.Fprint(writer, "Token is not valid")
				}
			} else {
				writer.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(writer, "Unauthorized access to this resource")
			}

		})
	}

	return nil
}
