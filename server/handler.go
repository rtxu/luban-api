package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/db"
	"github.com/rtxu/luban-api/middleware"
)

/**
该 API 的特殊性
「取数」类 API 即使失败，并不会影响前端页面的整体展示，前端会负责 error handling 逻辑，给用户提供反馈
该 API 拥有者用户页面的控制权，用户正常登陆流程如下：

1. 在 client 端发起登录请求，client 端将页面 redirect 给 GitHub
2. GitHub 授权成功后，redirect 到该 API
3. 该 API 对已授权用户进行处理（签发 token 等），redirect 给 client

如果该 API 在第三步出错，而未能 redirect 给 client，将会导致用户页面停留在 API Server 端，
打断了用户的使用，所以该 API 无论成功与否，都应该将控制权交给 client
*/
const kTokenClaimUserId = "user_id"
const kTokenClaimUserName = "username"

func (s *server) handleGithubLogin() http.HandlerFunc {
	type accessTokenT struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	type userInfoT struct {
		Login     string `json:"login"`
		AvatarUrl string `json:"avatar_url"`
	}
	const ACCESS_TOKEN_URL = "https://github.com/login/oauth/access_token"
	const USER_INFO_URL = "https://api.github.com/user"

	defaultRootDir := json.RawMessage("[]")

	loginErr := func(w http.ResponseWriter, r *http.Request, err error, promptErrMsg string) {
		middleware.GetLogEntry(r).Warningln(err)

		params := make(url.Values)
		params.Add("loginError", promptErrMsg)
		http.Redirect(w, r, fmt.Sprintf("%s/login?%s", s.conf.AppRoot, params.Encode()),
			http.StatusSeeOther)
		return
	}
	return func(w http.ResponseWriter, r *http.Request) {
		githubCode := r.URL.Query().Get("code")
		params := make(url.Values)
		params.Add("code", githubCode)
		params.Add("client_id", s.conf.GithubOAuth.ClientID)
		params.Add("client_secret", s.conf.GithubOAuth.ClientSecret)
		req, _ := http.NewRequest("POST", ACCESS_TOKEN_URL+"?"+params.Encode(), nil)
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			loginErr(w, r,
				fmt.Errorf("error happened when POST %s, err: %w", ACCESS_TOKEN_URL, err),
				"GitHub 登录出错")
			return
		}
		defer resp.Body.Close()

		var accessToken accessTokenT
		if err := json.NewDecoder(resp.Body).Decode(&accessToken); err != nil {
			loginErr(w, r,
				fmt.Errorf("error happened when decode(accessToken), err: %w", err),
				"GitHub 登录出错")
			return
		}

		req2, _ := http.NewRequest("GET", USER_INFO_URL, nil)
		req2.Header.Add("Authorization", fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken))
		resp2, err := client.Do(req2)
		if err != nil {
			loginErr(w, r,
				fmt.Errorf("error happened when GET %s, err: %w", USER_INFO_URL, err),
				"GitHub 登录出错")
			return
		}
		defer resp2.Body.Close()
		var userInfo userInfoT
		if err := json.NewDecoder(resp2.Body).Decode(&userInfo); err != nil {
			loginErr(w, r,
				fmt.Errorf("error happened when decode(userInfo), err: %w", err),
				"GitHub 登录出错")
			return
		}

		var user db.User
		user, err = s.userService.FindByGithubUserName(userInfo.Login)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				user = db.User{
					UserName:       userInfo.Login,
					GithubUserName: &userInfo.Login,
					AvatarUrl:      &userInfo.AvatarUrl,
					RootDir:        defaultRootDir,
				}
				err := s.userService.NewUser(&user)
				if err != nil {
					loginErr(w, r,
						fmt.Errorf("error happened when insert user, err: %w", err),
						"登录出错，请联系客服")
					return
				}
			} else {
				loginErr(w, r,
					fmt.Errorf("error happened when FindByGithubUserName(%s), err: %w",
						userInfo.Login, err),
					"登录出错，请联系客服")
				return
			}
		}

		claims := jwt.MapClaims{
			kTokenClaimUserId:   user.ID,
			kTokenClaimUserName: userInfo.Login,
		}
		jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
		_, tokenString, _ := s.tokenAuth.Encode(claims)
		query := url.Values{
			"access_token": {fmt.Sprintf("Bearer %s", tokenString)},
		}.Encode()

		http.Redirect(w, r,
			fmt.Sprintf("%s/login-success?%s", s.conf.AppRoot, query),
			http.StatusSeeOther)
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
		var username = claims[kTokenClaimUserName].(string)
		user, rootDir := s.getCurrentUserAndRootDirFromDB(username)
		data := dataT{
			Username:  user.UserName,
			AvatarUrl: *user.AvatarUrl,
			RootDir:   rootDir,
		}
		s.respond(w, r, defaultResponse{Data: data}, http.StatusOK)
	}
}
