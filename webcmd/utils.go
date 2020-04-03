package webcmd

import (
	"fmt"
	"reflect"
	"strings"
)

func firstLower(s string) string {
	ret := ""
	if s == "" {
		fmt.Printf("error " + s)
		return ""
	}
	if len(s) >= 1 {

		ret += strings.ToLower(string(s[0]))
		if len(s) > 1 {
			ret += s[1:]
		}
	}
	return ret
}

func getTagName(tagValue string) string {
	structTag := reflect.StructTag(strings.Replace(tagValue, "`", "", -1))
	jsonTag := structTag.Get("json")
	// json:"tag,hoge"
	if strings.Contains(jsonTag, ",") {
		// json:",hoge"
		if strings.HasPrefix(jsonTag, ",") {
			jsonTag = ""
		} else {
			jsonTag = strings.SplitN(jsonTag, ",", 2)[0]
		}
	}
	if jsonTag == "-" {
		return ""
	} else if jsonTag != "" {
		return jsonTag
	} else {
		return ""
	}
}
