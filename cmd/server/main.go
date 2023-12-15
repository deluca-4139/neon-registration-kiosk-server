package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Event struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StartDate   string `json:"startDate"`
	StartTime   string `json:"startTime"`
	EndDate     string `json:"endDate"`
	EndTime     string `json:"endTime"`
}

type Pagination struct {
	CurrentPage  int32 `json:"currentPage"`
	PageSize     int32 `json:"pageSize"`
	TotalPages   int32 `json:"totalPages"`
	TotalResults int32 `json:"totalResults"`
}

type EventRequest struct {
	Events     []Event    `json:"events"`
	Pagination Pagination `json:"pagination"`
}

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
	q.Set("startDateAfter", time.Now().Format(time.DateOnly))
	q.Set("startDateBefore", time.Now().Add(time.Hour*24).Format(time.DateOnly))
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Add("NEON-API-VERSION", "2.6")
	auth_string := []byte(fmt.Sprintf("orgId:%v", neonKey))
	encoded_auth := base64.StdEncoding.EncodeToString(auth_string)
	req.Header.Add("Authorization", "Basic "+encoded_auth)

	resp, err := client.Do(req)

	var msg EventRequest
	var decResp []byte

	if err != nil {
		// do something to fix error
	} else if resp.StatusCode != 200 {
		// do something to note non-200 response
	} else {
		json.NewDecoder(resp.Body).Decode(&msg)
		decResp, _ = json.Marshal(msg)
		fmt.Printf(string(decResp))
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(decResp)
}
