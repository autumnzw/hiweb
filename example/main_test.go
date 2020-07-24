package main

import (
	"testing"

	"github.com/autumnzw/hiweb/webcmd"
)

func TestAutoRoute(t *testing.T) {
	err := webcmd.CreateRoute("./controllers", "example", "", "")
	if err != nil {
		t.Error(err)
	}

}
