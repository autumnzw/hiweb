package webcmd

import (
	"fmt"
	"testing"
)

func TestCreateRoute(t *testing.T) {
	err := CreateRoute("./controllers", "hiweb", "http://localhost:8080", "./controllers/api.js")
	fmt.Printf("err:%s", err)
}
