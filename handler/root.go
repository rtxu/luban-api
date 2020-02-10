package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	"github.com/sirupsen/logrus"

	"github.com/rtxu/luban-api/config"
	"github.com/rtxu/luban-api/db"
)

var tokenAuth *jwtauth.JWTAuth = jwtauth.New("HS256", []byte(config.JWTSecret), nil)
var Root http.Handler

func handleGithubLogin(w http.ResponseWriter, r *http.Request) {
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
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	type accessTokenT struct {
		TokenType   string `json:"token_type"`
		AccessToken string `json:"access_token"`
	}
	var accessToken accessTokenT
	if err := json.Unmarshal(body, &accessToken); err != nil {
		panic(err)
	}

	req2, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req2.Header.Add("Authorization", fmt.Sprintf("%s %s", accessToken.TokenType, accessToken.AccessToken))
	resp2, err := client.Do(req2)
	if err != nil {
		panic(err)
	}
	defer resp2.Body.Close()
	body2, err := ioutil.ReadAll(resp2.Body)
	type userInfoT struct {
		Login     string `json:"login"`
		AvatarUrl string `json:"avatar_url"`
	}
	var userInfo userInfoT
	if err := json.Unmarshal(body2, &userInfo); err != nil {
		panic(err)
	}

	_, err = db.UserService.FindByGithubUserName(userInfo.Login)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err := db.UserService.Insert(db.User{
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
	_, tokenString, _ := tokenAuth.Encode(claims)
	query := url.Values{
		"access_token": {fmt.Sprintf("Bearer %s", tokenString)},
	}.Encode()

	http.Redirect(w, r,
		fmt.Sprintf("%s/login-success?%s", config.AppRoot, query),
		303)
}

type JsonResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func init() {

	// Setup the logger backend using sirupsen/logrus and configure
	// it to use a custom JSONFormatter. See the logrus docs for how to
	// configure the backend at github.com/sirupsen/logrus
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}

	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	r.Use(middleware.RequestID)
	r.Use(NewStructuredLogger(logger))
	r.Use(middleware.Recoverer)

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("welcome"))
		})

		r.Get("/callback/github/login", handleGithubLogin)
		r.Get("/callback/github/signup", handleGithubLogin)
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
		r.Get("/currentUser", func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			fmt.Println()
			var userID = claims["user_id"].(string)
			user, err := db.UserService.Find(userID)
			if err != nil {
				panic(err)
			}
			var jsonResponse = JsonResponse{
				Code: 0,
				Msg:  "",
				Data: user,
			}
			jsonBytes, _ := json.Marshal(jsonResponse)
			w.Write(jsonBytes)
		})
	})

	Root = r
}

// StructuredLogger is a simple, but powerful implementation of a custom structured
// logger backed on logrus. I encourage users to copy it, adapt it and make it their
// own. Also take a look at https://github.com/pressly/lg for a dedicated pkg based
// on this work, designed for context-based http routers.

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

type StructuredLogger struct {
	Logger *logrus.Logger
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.Logger = entry.Logger.WithFields(logFields)

	entry.Logger.Infoln("request started")

	return entry
}

type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.
//
// This is a useful pattern to use to set state on the entry as it
// passes through the handler chain, which at any point can be logged
// with a call to .Print(), .Info(), etc.

func GetLogEntry(r *http.Request) logrus.FieldLogger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.Logger
}

func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithField(key, value)
	}
}

func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
		entry.Logger = entry.Logger.WithFields(fields)
	}
}
