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

// Copied from the [golang official doc](https://golang.org/pkg/net/http/#ResponseWriter):
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
// The provided code must be a valid HTTP 1xx-5xx status code.
// Only one header may be written.

// 经测试，此处 WriterHeader 后，如果发生 panic，返回的 reponse header 依然是 200，
// 而非 500（由 middleware Recoverer 实现）
func (s *server) respond(w http.ResponseWriter, r *http.Request, data interface{}, status int) {
	var resp *defaultResponse
	switch v := data.(type) {
	case defaultResponse:
		resp = &v
	case error:
		if errCode, ok := errCodeMap[errors.Unwrap(v)]; ok {
			resp = &defaultResponse{
				Code: errCode,
				Msg:  v.Error(),
			}
		} else {
			panic(fmt.Sprintf("unexpected error: %v", v))
		}

	default:
		panic(fmt.Sprintf("unknown data type: %T", data))
	}

	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		panic(err)
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
