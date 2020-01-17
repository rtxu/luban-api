package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/config"
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(config.JWTSecret), nil)
}

func httpRouter() http.Handler {
	r := chi.NewRouter()

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})

		r.Get("/callback/github/login", func(w http.ResponseWriter, r *http.Request) {
			githubCode := r.URL.Query().Get("code")
			params := make(url.Values)
			params.Add("code", githubCode)
			params.Add("client_id", config.GithubOAuth.ClientID)
			params.Add("client_secret", config.GithubOAuth.ClientSecret)
			req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token?"+params.Encode(), nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// TODO
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			type accessTokenT struct {
				TokenType   string `json:"token_type"`
				AccessToken string `json:"access_token"`
			}
			var accessToken accessTokenT
			json.Unmarshal(body, &accessToken)

			req2, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
			req2.Header.Add("Authorization", fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken))
			resp2, err := client.Do(req2)
			if err != nil {
				// TODO
			}
			defer resp2.Body.Close()
			body2, err := ioutil.ReadAll(resp2.Body)
			type userInfoT struct {
				Login string `json:"login"`
			}
			var userInfo userInfoT
			json.Unmarshal(body2, &userInfo)

			claims := jwt.MapClaims{
				"user_id": userInfo.Login,
			}
			jwtauth.SetExpiryIn(claims, 7*24*time.Hour)
			_, tokenString, _ := tokenAuth.Encode(claims)
			query := url.Values{
				"access_token": {fmt.Sprintf("Bearer %s", tokenString)},
			}.Encode()

			http.Redirect(w, r,
				fmt.Sprintf("%s/login-success?%s", config.AppRoot, query),
				303)
		})
	})

	// Protected Routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(tokenAuth))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator)

		r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("protected area. hi %v", claims["user_id"])))
		})
	})

	return r
}

func main() {
	http.ListenAndServe(":9090", httpRouter())
}
