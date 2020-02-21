package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/db"
)

func (s *server) handleGithubLogin() http.HandlerFunc{
	type accessTokenT struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	type userInfoT struct {
		Login     string `json:"login"`
		AvatarUrl string `json:"avatar_url"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		githubCode := r.URL.Query().Get("code")
		params := make(url.Values)
		params.Add("code", githubCode)
		params.Add("client_id", s.conf.GithubOAuth.ClientID)
		params.Add("client_secret", s.conf.GithubOAuth.ClientSecret)
		req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token?"+params.Encode(), nil)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		var accessToken accessTokenT
		if err := json.NewDecoder(resp.Body).Decode(&accessToken); err != nil {
			panic(err)
		}

		req2, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
		req2.Header.Add("Authorization", fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken))
		resp2, err := client.Do(req2)
		if err != nil {
			panic(err)
		}
		defer resp2.Body.Close()
		var userInfo userInfoT
		if err := json.NewDecoder(resp2.Body).Decode(&userInfo); err != nil {
			panic(err)
		}

		_, err = s.userService.FindByGithubUserName(userInfo.Login)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				err := s.userService.Insert(db.User{
					UserName:       userInfo.Login,
					GithubUserName: &userInfo.Login,
					AvatarUrl:      &userInfo.AvatarUrl,
				})
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		claims := jwt.MapClaims{
			"user_id": userInfo.Login,
		}
		jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
		_, tokenString, _ := s.tokenAuth.Encode(claims)
		query := url.Values{
			"access_token": {fmt.Sprintf("Bearer %s", tokenString)},
		}.Encode()

		http.Redirect(w, r,
			fmt.Sprintf("%s/login-success?%s", s.conf.AppRoot, query),
			303)
	}
}

func (s *server) handleCurrentUserGet() http.HandlerFunc {
	type dataT struct {
		Username  string     `json:"username"`
		AvatarUrl string     `json:"avatarUrl"`
		RootDir   DirectoryT `json:"rootDir"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		_, claims, _ := jwtauth.FromContext(r.Context())
		var userID = claims["user_id"].(string)
		user, err := s.userService.Find(userID)
		if err != nil {
			panic(err)
		}
		data := dataT{
			Username:  user.UserName,
			AvatarUrl: *user.AvatarUrl,
		}
		if user.RootDir != nil {
			err := json.NewDecoder(strings.NewReader(*user.RootDir)).Decode(&data.RootDir)
			if err != nil {
				panic(err)
			}
		}
		s.respond(w, r, defaultResponse{Data: data}, http.StatusOK)
	}
}