package main

import (
	"net/http"

	"github.com/rtxu/luban-api/handler"
)

func main() {
	http.ListenAndServe(":9090", handler.Root)
}
