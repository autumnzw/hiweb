package main

import (
	_ "example/controllers"
	"fmt"

	"flag"
	"net/http"

	"github.com/autumnzw/hiweb"
)

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip")
	port := flag.String("port", "8111", "port")
	flag.Parse()
	if *port == "" {
		flag.PrintDefaults()
		return
	}
	hiweb.WebConfig.SecretKey = "asdfsvasf"
	hiweb.RouteFiles("/", "./dist")
	fmt.Printf("start web %s:%s dir:%s", *ip, *port, "./dist")
	e := http.ListenAndServe(*ip+":"+*port, nil)
	if e != nil {
		fmt.Printf("err:%s", e)
	}
}
