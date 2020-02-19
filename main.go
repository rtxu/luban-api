package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"

	"github.com/rtxu/luban-api/handler"
	myMiddleware "github.com/rtxu/luban-api/middleware"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()

	// [Setup global middleware

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
	r.Use(myMiddleware.NewStructuredLogger())
	r.Use(middleware.Recoverer)

	// Setup global middleware]

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Get("/callback/github/login", handler.GithubLogin)
		r.Get("/callback/github/signup", handler.GithubLogin)
	})

	// Protected Routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtauth.Verifier(handler.TokenAuth))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		r.Use(jwtauth.Authenticator)

		r.Get("/currentUser", handler.GetCurrentUser)

		r.Route("/currentUser/entry", func(r chi.Router) {
			r.Post("/", handler.AddEntry)
			r.Delete("/", handler.DeleteEntry)
		})
	})

	return r
}

func main() {
	log.Fatal(http.ListenAndServe(":9090", NewRouter()))
}
