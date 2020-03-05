package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleApp(t *testing.T) {
	assert := assert.New(t)
	svr, token := newTestServer()

	// CREATE
	createEntry(createRequest{
		Dir: "/",
		Entry: EntryT{
			Name: "entry1",
			Type: App,
		},
	}, svr, token)

	// PUT, publish or save
	publishContent := make(map[string]interface{})
	saveContent := make(map[string]interface{})
	{
		query := make(url.Values)
		query.Add("appId", fmt.Sprintf("%d", 0))

		{
			query.Set("op", kOpPublish)
			publishContent["widgets"] = map[string]interface{}{
				"published_w1": map[string]interface{}{},
			}
			contentBytes, _ := json.Marshal(publishContent)
			httpReq := httptest.NewRequest("PUT", "/currentUser/app?"+query.Encode(),
				bytes.NewReader(contentBytes))
			httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
			resp := handleRequest(httpReq, svr)

			assertErrCode(t, success.Code, resp)
		}

		{
			query.Set("op", kOpSave)
			saveContent["widgets"] = map[string]interface{}{
				"save_w1": map[string]interface{}{},
			}
			contentBytes, _ := json.Marshal(saveContent)
			httpReq := httptest.NewRequest("PUT", "/currentUser/app?"+query.Encode(),
				bytes.NewReader(contentBytes))
			httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
			resp := handleRequest(httpReq, svr)

			assertErrCode(t, success.Code, resp)
		}
	}

	// GET, view/preview/edit
	{
		query := make(url.Values)
		query.Add("appId", fmt.Sprintf("%d", 0))
		{
			query.Set("loadType", kLTView)
			httpReq := httptest.NewRequest("GET", "/currentUser/app?"+query.Encode(), nil)
			httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
			resp := handleRequest(httpReq, svr)

			jsonResponse := assertErrCode(t, success.Code, resp)
			assert.Equal(publishContent, jsonResponse.Data)
		}
		{
			for _, loadType := range []string{kLTEdit, kLTPreview} {
				query.Set("loadType", loadType)
				httpReq := httptest.NewRequest("GET", "/currentUser/app?"+query.Encode(), nil)
				httpReq.Header.Add("Authorization", fmt.Sprintf("BEARER %s", token))
				resp := handleRequest(httpReq, svr)

				jsonResponse := assertErrCode(t, success.Code, resp)
				assert.Equal(saveContent, jsonResponse.Data)
			}
		}
	}
}
