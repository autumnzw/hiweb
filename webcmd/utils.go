package webcmd

import (
	"fmt"
	"go/format"
	"os"
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

// PathExists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err

}

// FormatSource
func FormatSource(src []byte) []byte {
	code, err := format.Source(src)
	if err != nil {
		code = src // Output the unformated code anyway
	}
	return code
}
