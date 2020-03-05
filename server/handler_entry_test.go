package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	myCreateEntry := func(req createRequest) *http.Response {
		return createEntry(req, svr, token)
	}
	assert := assert.New(t)
	assertEntryExists := func(entryName string, dir DirectoryT) {
		for _, entry := range dir {
			if entry.Name == entryName {
				if entry.Type == App {
					_, err := svr.appService.Find(kTestUserId, entry.AppId)
					assert.NoError(err)
				}
				return
			}
		}
		t.Errorf("NOT FOUND entry: %s", entryName)
	}

	// 1. create app /entry1
	resp := myCreateEntry(createRequest{
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

	_, rootDir := svr.getCurrentUserAndRootDirFromDB(kTestUserName)
	assertEntryExists("entry1", rootDir)

	// 2. create app /entry1 again
	assertErrCode(t, errCodeMap[errEntryAlreadyExist], myCreateEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "entry1",
			Type: App,
		},
	}))

	// 3. create app /a/b/c
	assertErrCode(t, success.Code, myCreateEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "a",
			Type: Directory,
		},
	}))
	assertErrCode(t, success.Code, myCreateEntry(createRequest{
		Dir: "/a/",
		Entry: EntryT{
			Name: "b",
			Type: Directory,
		},
	}))
	assertErrCode(t, success.Code, myCreateEntry(createRequest{
		Dir: "/a/b/",
		Entry: EntryT{
			Name: "c",
			Type: App,
		},
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kTestUserName)
	dir, err := findDir("/a/b/", &rootDir)
	assert.Nil(err)
	assertEntryExists("c", *dir)

	// 4. invalid param
	// empty dir name
	assertErrCode(t, errCodeMap[errInvalidParam], myCreateEntry(createRequest{
		Dir: "",
		Entry: EntryT{
			Name: "c",
			Type: App,
		},
	}))
	// not begin with '/'
	assertErrCode(t, errCodeMap[errInvalidParam], myCreateEntry(createRequest{
		Dir: "dir/",
		Entry: EntryT{
			Name: "entry",
			Type: App,
		},
	}))
	// not end with '/'
	assertErrCode(t, errCodeMap[errInvalidParam], myCreateEntry(createRequest{
		Dir: "/dir",
		Entry: EntryT{
			Name: "entry",
			Type: App,
		},
	}))
	// empty entry name
	assertErrCode(t, errCodeMap[errInvalidParam], myCreateEntry(createRequest{
		Dir: "/dir/",
		Entry: EntryT{
			Name: "",
			Type: App,
		},
	}))
	// entry name contains '/'
	assertErrCode(t, errCodeMap[errInvalidParam], myCreateEntry(createRequest{
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
	assertErrCode(t, success.Code, deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "not_exist_entry",
	}))

	// 2. delete /entry1
	assertErrCode(t, success.Code, deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "entry1",
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kTestUserName)
	assertEntryNotExist("entry1", rootDir)

	// 3. delete non-empty directory
	assertErrCode(t, errCodeMap[errDirNotEmpty], deleteEntry(deleteRequest{
		Dir:       "/",
		EntryName: "a",
	}))

	// 4. delete empty directory
	assertErrCode(t, success.Code, deleteEntry(deleteRequest{
		Dir:       "/a/b/",
		EntryName: "c",
	}))
	assertErrCode(t, success.Code, deleteEntry(deleteRequest{
		Dir:       "/a/",
		EntryName: "b",
	}))
	_, rootDir = svr.getCurrentUserAndRootDirFromDB(kTestUserName)
	dir, err = findDir("/a/", &rootDir)
	assert.Nil(err)
	assertEntryNotExist("b", *dir)
}
