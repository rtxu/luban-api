package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/rtxu/luban-api/config"
	"github.com/rtxu/luban-api/db"
	"upper.io/db.v3/lib/sqlbuilder"
)

// refactor based on [GopherCon 2019: Mat Ryer - How I Write HTTP Web Services after Eight Years](https://www.youtube.com/watch?v=rWBSMsLG8po)
type server struct {
	conf      config.AppConfig
	router    chi.Router
	tokenAuth *jwtauth.JWTAuth

	appService  db.AppService
	userService db.UserService
}

func New(conf config.AppConfig) *server {
	svr := &server{
		conf:      conf,
		router:    chi.NewRouter(),
		tokenAuth: jwtauth.New("HS256", []byte(conf.JWTSecret), nil),
	}
	svr.routes()
	return svr
}

func (s *server) SetupDBService(dbConn sqlbuilder.Database) {
	s.appService = db.NewAppService(dbConn)
	s.userService = db.NewUserService(dbConn)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	w.WriteHeader(status)
	if data != nil {
		var resp *defaultResponse
		switch v := data.(type) {
		case defaultResponse:
			resp = &v
		case error:
			resp = &defaultResponse{
				Code: errCodeMap[errors.Unwrap(v)],
				Msg:  v.Error(),
			}
		default:
			panic(fmt.Sprintf("unknown data type: %T", data))
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
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

var success = defaultResponse{}
