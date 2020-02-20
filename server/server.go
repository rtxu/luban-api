package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"upper.io/db.v3/lib/sqlbuilder"
)

// refactor based on [GopherCon 2019: Mat Ryer - How I Write HTTP Web Services after Eight Years](https://www.youtube.com/watch?v=rWBSMsLG8po)
type server struct {
	db     sqlbuilder.Database
	router chi.Router
}

func New() *server {
	svr := &server{
		router: chi.NewRouter(),
	}
	svr.routes()
	return svr
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			panic(err)
		}
	}
}

func (s *server) decode(w http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// 返回给用户的结果均遵循该结构
type defaultResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
