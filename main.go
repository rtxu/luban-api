package main

import (
	"log"
	"net/http"

	"github.com/rtxu/luban-api/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	svr := server.New()
	return http.ListenAndServe(":9090", svr)
}
