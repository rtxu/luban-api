package server

import (
	"bytes"
	"encoding/json"
	"errors"
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

const kCurrentUserName = "test_user"

func newTestServer() (*server, string) {
	conf := config.AppConfig{
		JWTSecret: "test_secret",
	}
	svr := New(conf)

	svr.appService = db.NewMemAppService()
	svr.userService = db.NewMemUserService()
	svr.userService.Insert(db.User{
		UserName: kCurrentUserName,
	})
	claims := jwt.MapClaims{
		"user_id": kCurrentUserName,
	}
	jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
	_, tokenString, _ := svr.tokenAuth.Encode(claims)
	return svr, tokenString
}

func TestFindDir(t *testing.T) {
	assert := assert.New(t)
	{
		root := &DirectoryT{}
		ptr, err := findDir("/", root)
		assert.Nil(err)
		assert.Equal(root, ptr)
	}
	{
		root := &DirectoryT{}
		_, err := findDir("/not_exist_dir/", root)
		assert.True(errors.Is(err, errEntryNotFound))
	}
	{
		targetDir := DirectoryT{}
		root := &DirectoryT{
			&EntryT{Name: "a", Type: Directory, Children: DirectoryT{
				&EntryT{Name: "b", Type: Directory, Children: targetDir},
			}},
		}
		ptr, err := findDir("/a/b/", root)
		assert.Nil(err)
		assert.Equal(&targetDir, ptr)
	}
}

func TestHandleEntry(t *testing.T) {
	svr, token := newTestServer()

	// POST /currentUser/entry
	type createRequest struct {
		Dir   string `json:"dir"`
		Entry EntryT `json:"entry"`
	}
	createEntry := func(req createRequest) *http.Response {
		reqBodyBytes, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/currentUser/entry",
			bytes.NewReader(reqBodyBytes))
		httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
		// fmt.Println(req.Header.Get("Authorization"))
		w := httptest.NewRecorder()
		svr.ServeHTTP(w, httpReq)
		return w.Result()
	}

	assert := assert.New(t)
	assertEntryExists := func(entryName string, dir DirectoryT) {
		for _, entry := range dir {
			if entry.Name == entryName {
				return
			}
		}
		t.Errorf("NOT FOUND entry: %s", entryName)
	}
	assertErrCode := func(code int, resp *http.Response) {
		assert.Equal(http.StatusOK, resp.StatusCode)

		var jsonResponse defaultResponse
		json.NewDecoder(resp.Body).Decode(&jsonResponse)
		assert.Equal(code, jsonResponse.Code)
	}

	// 1. create app /entry1
	resp := createEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "entry1",
			Type: App,
		},
	})
	assert.Equal(http.StatusOK, resp.StatusCode)

	var jsonResponse defaultResponse
	json.NewDecoder(resp.Body).Decode(&jsonResponse)
	assert.Equal(0, jsonResponse.Code)

	_, rootDir := svr.getCurrentUserAndRootDirFromDB(kCurrentUserName)
	assertEntryExists("entry1", rootDir)

	// 2. create app /entry1 again
	assertErrCode(errCodeMap[errEntryAlreadyExist], createEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "entry1",
			Type: App,
		},
	}))

	// 3. create app /a/b/c
	assertErrCode(success.Code, createEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "a",
			Type: Directory,
		},
	}))
	assertErrCode(success.Code, createEntry(createRequest{
		Dir: "/a/",
		Entry: EntryT{
			Name: "b",
			Type: Directory,
		},
	}))
	assertErrCode(success.Code, createEntry(createRequest{
		Dir: "/a/b/",
		Entry: EntryT{
			Name: "c",
			Type: App,
		},
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kCurrentUserName)
	dir, err := findDir("/a/b/", &rootDir)
	assert.Nil(err)
	assertEntryExists("c", *dir)

	// 4. invalid param
	// empty dir name
	assertErrCode(errCodeMap[errInvalidParam], createEntry(createRequest{
		Dir: "",
		Entry: EntryT{
			Name: "c",
			Type: App,
		},
	}))
	// not begin with '/'
	assertErrCode(errCodeMap[errInvalidParam], createEntry(createRequest{
		Dir: "dir/",
		Entry: EntryT{
			Name: "entry",
			Type: App,
		},
	}))
	// not end with '/'
	assertErrCode(errCodeMap[errInvalidParam], createEntry(createRequest{
		Dir: "/dir",
		Entry: EntryT{
			Name: "entry",
			Type: App,
		},
	}))
	// empty entry name
	assertErrCode(errCodeMap[errInvalidParam], createEntry(createRequest{
		Dir: "/dir/",
		Entry: EntryT{
			Name: "",
			Type: App,
		},
	}))
	// entry name contains '/'
	assertErrCode(errCodeMap[errInvalidParam], createEntry(createRequest{
		Dir: "/dir/",
		Entry: EntryT{
			Name: "entry_name_contains_/",
			Type: App,
		},
	}))

	// DELETE /currentUser/entry
	type deleteRequest struct {
		Dir       string `json:"dir"`
		EntryName string `json:"entryName"`
	}
	deleteEntry := func(req deleteRequest) *http.Response {
		reqBodyBytes, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("DELETE", "/currentUser/entry",
			bytes.NewReader(reqBodyBytes))
		httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
		w := httptest.NewRecorder()
		svr.ServeHTTP(w, httpReq)
		return w.Result()
	}
	assertEntryNotExist := func(entryName string, dir DirectoryT) {
		for _, entry := range dir {
			if entry.Name == entryName {
				t.Errorf("FOUND entry: %s", entryName)
				return
			}
		}
	}

	// 1. delete not exist entry
	assertErrCode(success.Code, deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "not_exist_entry",
	}))

	// 2. delete /entry1
	assertErrCode(success.Code, deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "entry1",
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kCurrentUserName)
	assertEntryNotExist("entry1", rootDir)

	// 3. delete non-empty directory
	assertErrCode(errCodeMap[errDirNotEmpty], deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "a",
	}))

	// 4. delete empty directory
	assertErrCode(success.Code, deleteEntry(deleteRequest{
		Dir:       "/a/b/",
		EntryName: "c",
	}))
	assertErrCode(success.Code, deleteEntry(deleteRequest{
		Dir:       "/a/",
		EntryName: "b",
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kCurrentUserName)
	dir, err = findDir("/a/", &rootDir)
	assert.Nil(err)
	assertEntryNotExist("b", *dir)
}
