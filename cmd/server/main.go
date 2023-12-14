package main

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	// "strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

var client = &http.Client{}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

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
	req, _ := http.NewRequest("GET", "https://api.neoncrm.com/v2/events", nil)
	req.Header.Add("NEON-API-VERSION", "2.6")
	auth_string := []byte("orgId:secret_key")
	encoded_auth := base64.StdEncoding.EncodeToString(auth_string)
	req.Header.Add("Authorization", "Basic "+encoded_auth)

	resp, _ := client.Do(req)

	w.WriteHeader(resp.StatusCode)
	b, _ := ioutil.ReadAll(resp.Body)
	w.Write(b)
}
