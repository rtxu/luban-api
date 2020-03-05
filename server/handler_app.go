package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth"
)

func (s *server) handleAppGet() http.HandlerFunc {
	type request struct {
		appId    string
		loadType string
	}
	const (
		LT_View    = "view"
		LT_Preview = "preview"
		LT_Edit    = "edit"
	)
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var owner = uint32(claims[kTokenClaimUserId].(float64))

		var param request
		query := r.URL.Query()
		param.appId = query.Get("appId")
		param.loadType = query.Get("loadType")

		u64, err := strconv.ParseUint(param.appId, 10, 32)
		if err != nil {
			s.respond(w, r, fmt.Errorf("%w: appId(%s) is not a number",
				errBadRequest, param.appId), http.StatusOK)
			return
		}
		appId := uint32(u64)
		app, err := s.appService.Find(owner, appId)
		if err != nil {
			s.respond(w, r, fmt.Errorf("%w: appId is %s",
				errEntryNotFound, param.appId), http.StatusOK)
			return
		}

		var content json.RawMessage
		switch param.loadType {
		case LT_View:
			content = app.LastPublishedContent
		case LT_Edit, LT_Preview:
			content = app.Content
		default:
			s.respond(w, r, fmt.Errorf("%w: unrecognized loadType(%s)",
				errBadRequest, param.loadType), http.StatusOK)
			return
		}

		fmt.Printf("load app success, appId: %d, loadType: %s, content: %s", appId, param.loadType, string(content))
		s.respond(w, r, defaultResponse{
			Data: content,
		}, http.StatusOK)
	}
}
