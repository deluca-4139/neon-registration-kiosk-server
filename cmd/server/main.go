package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var client = &http.Client{}

func main() {
	rootCmd.Execute()
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	// Based on path of execution, not path of file
	http.ServeFile(w, r, "web/static/landing.html")
}

func verifyRegistration(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse("https://api.neoncrm.com/v2/events")
	q := u.Query()
	q.Set("startDateAfter", "2023-11-01")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("NEON-API-VERSION", "2.6")
	auth_string := []byte(fmt.Sprintf("orgId:%v", neonKey))
	encoded_auth := base64.StdEncoding.EncodeToString(auth_string)
	req.Header.Add("Authorization", "Basic "+encoded_auth)

	resp, _ := client.Do(req)

	w.WriteHeader(resp.StatusCode)
	b, _ := ioutil.ReadAll(resp.Body)
	w.Write(b)
}
