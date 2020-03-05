package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/rtxu/luban-api/config"
	"github.com/rtxu/luban-api/db"
	"github.com/stretchr/testify/assert"
)

const kTestUserName = "test_user"
const kTestUserId = 9527

func newTestServer() (*server, string) {
	conf := config.AppConfig{
		JWTSecret: "test_secret",
	}
	svr := New(conf)

	svr.appService = db.NewMemAppService()
	svr.userService = db.NewMemUserService()
	svr.userService.Insert(db.User{
		UserName: kTestUserName,
		ID:       kTestUserId,
	})
	claims := jwt.MapClaims{
		kTokenClaimUserName: kTestUserName,
		kTokenClaimUserId:   kTestUserId,
	}
	jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
	_, tokenString, _ := svr.tokenAuth.Encode(claims)
	return svr, tokenString
}

func assertErrCode(t *testing.T, code int, resp *http.Response) defaultResponse {
	assert := assert.New(t)
	assert.Equal(http.StatusOK, resp.StatusCode)

	var jsonResponse defaultResponse
	json.NewDecoder(resp.Body).Decode(&jsonResponse)
	assert.Equal(code, jsonResponse.Code)

	return jsonResponse
}

func handleRequest(req *http.Request, svr *server) *http.Response {
	w := httptest.NewRecorder()
	svr.ServeHTTP(w, req)
	return w.Result()
}

type createRequest struct {
	Dir   string `json:"dir"`
	Entry EntryT `json:"entry"`
}

func createEntry(req createRequest, svr *server, token string) *http.Response {
	reqBodyBytes, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/currentUser/entry",
		bytes.NewReader(reqBodyBytes))
	httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
	// fmt.Println(req.Header.Get("Authorization"))
	w := httptest.NewRecorder()
	svr.ServeHTTP(w, httpReq)
	return w.Result()
}
