package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", landingPage)
	r.Get("/form.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/form.js")
	})
	r.Post("/verify", verifyRegistration)

	http.ListenAndServe(":3000", r)
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	// Based on path of execution, not path of file
	http.ServeFile(w, r, "web/static/landing.html")
}

func verifyRegistration(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("registration valid!"))
}
