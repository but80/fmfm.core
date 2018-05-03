package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

func init() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	addr := strings.Split(l.Addr().String(), ":")
	port := addr[len(addr)-1]
	fmt.Printf("> go tool pprof http://127.0.0.1:%s/debug/pprof/profile\n", port)
	fmt.Printf("> pprof -http=localhost:8080 ~/pprof/pprof.127.0.0.1:%s.samples.cpu.001.pb.gz\n", port)
	go http.Serve(l, nil)
}
