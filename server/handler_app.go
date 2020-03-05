package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth"
)

const (
	kLTView    = "view"
	kLTPreview = "preview"
	kLTEdit    = "edit"
)

func (s *server) handleAppGet() http.HandlerFunc {
	type request struct {
		appId    string
		loadType string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var owner = uint32(claims[kTokenClaimUserId].(float64))

		var param request
		query := r.URL.Query()
		param.appId = query.Get("appId")
		param.loadType = query.Get("loadType")

		u64, err := strconv.ParseUint(param.appId, 10, 32)
		if err != nil {
			s.respond(w, r, fmt.Errorf("%w: appId(%s) is not a number, err: %v",
				errBadRequest, param.appId, err), http.StatusOK)
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
		case kLTView:
			content = app.LastPublishedContent
		case kLTEdit, kLTPreview:
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

const (
	kOpPublish = "publish"
	kOpSave    = "save"
)

func (s *server) handleAppSave() http.HandlerFunc {
	type request struct {
		appId string
		op    string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var owner = uint32(claims[kTokenClaimUserId].(float64))

		var param request
		query := r.URL.Query()
		param.appId = query.Get("appId")
		param.op = query.Get("op")

		u64, err := strconv.ParseUint(param.appId, 10, 32)
		if err != nil {
			s.respond(w, r, fmt.Errorf("%w: appId(%s) is not a number, err: %v",
				errBadRequest, param.appId, err), http.StatusOK)
			return
		}
		appId := uint32(u64)

		newContentBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.respond(w, r, fmt.Errorf("%w: failed to read body, err: %v",
				errBadRequest, err), http.StatusOK)
			return
		}

		switch param.op {
		case kOpPublish:
			err = s.appService.UpdateLastPublishedContent(owner, appId, newContentBytes)
		case kOpSave:
			err = s.appService.UpdateContent(owner, appId, newContentBytes)
		default:
			s.respond(w, r, fmt.Errorf("%w: unrecognized op(%s)",
				errBadRequest, param.op), http.StatusOK)
			return
		}

		if err == nil {
			s.respond(w, r, success, http.StatusOK)
		} else {
			panic(err)
		}
	}
}
