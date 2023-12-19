package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type IDNamePair struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

type CustomField struct {
	ID           string        `json:"id"`
	Value        string        `json:"value"`
	Name         string        `json:"name"`
	OptionValues *[]IDNamePair `json:"optionValues"`
}

type Event struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	// Description string `json:"description"`
	StartDate string `json:"startDate"`
	StartTime string `json:"startTime"`
	EndDate   string `json:"endDate"`
	EndTime   string `json:"endTime"`
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

type EventAttendee struct {
	AttendeeID           int32         `json:"attendeeId"`
	AccountID            string        `json:"accountId"`
	AttendeeCustomFields []CustomField `json:"attendeeCustomFields"`
}

type EventAttendees struct {
	Pagination Pagination      `json:"pagination"`
	Attendees  []EventAttendee `json:"attendees"`
}

type IndividualAccount struct {
	AccountID           string        `json:"accountId"`
	AccountCustomFields []CustomField `json:"accountCustomFields"`
}

type Account struct {
	IndividualAccount IndividualAccount `json:"individualAccount"`
}

var client = &http.Client{}

var eventAttendeesMap map[string]string

func main() {
	eventAttendeesMap = make(map[string]string)
	rootCmd.Execute()
}

// func landingPage(w http.ResponseWriter, r *http.Request) {
// 	// Based on path of execution, not path of file
// 	http.ServeFile(w, r, "web/static/landing.html")
// }

func makeNeonRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("NEON-API-VERSION", "2.6")
	auth_string := []byte(fmt.Sprintf("ordId:%v", neonKey))
	encoded_auth := base64.StdEncoding.EncodeToString(auth_string)

	req.Header.Set("Accept", "application/json")
	req.Header.Add("Authorization", "Basic "+encoded_auth)

	return client.Do(req)
}

func refreshEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse("https://api.neoncrm.com/v2/events")
	q := u.Query()
	q.Set("startDateAfter", time.Now().Format(time.DateOnly))
	// q.Set("startDateBefore", time.Now().Add(time.Hour*24).Format(time.DateOnly))
	u.RawQuery = q.Encode()

	resp, err := makeNeonRequest("GET", u.String(), nil)

	var msg EventRequest
	var firstEvent Event
	var decResp []byte

	if err != nil {
		// do something to fix error
	} else if resp.StatusCode != 200 {
		// do something to note non-200 response
	} else {
		json.NewDecoder(resp.Body).Decode(&msg)
		firstEvent = msg.Events[0] // TODO: what if there are more/less events?
		decResp, _ = json.Marshal(firstEvent)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(decResp)
	}
}

func updateAttendees(w http.ResponseWriter, r *http.Request) {
	// TODO: sanitize input of attendees by
	// upcasing ID values before putting
	// into attendee map

	// eventId := r.FormValue("eventId")
	eventId := 23091

	u, _ := url.Parse(fmt.Sprintf("https://api.neoncrm.com/v2/events/%v/attendees", eventId))
	q := u.Query()
	u.RawQuery = q.Encode()

	resp, _ := makeNeonRequest("GET", u.String(), nil)

	var msg EventAttendees

	json.NewDecoder(resp.Body).Decode(&msg)
	fmt.Printf("Total number of predicted entries: %v\n", msg.Pagination.TotalResults)

	// TODO: detect map collision?
	for { // can also format as while msg.Attendees != null
		for _, element := range msg.Attendees {

			// // TODO: possibly useful for determining
			// // guests and their ID info
			// var value string
			// var isMember bool
			// for _, field := range element.AttendeeCustomFields {
			// 	if field.ID == "91" {
			// 		if field.OptionValues == nil {
			// 			// TODO: fix error handling
			// 			continue
			// 		}
			// 		isMember = (*field.OptionValues)[0].ID == "17" // 17 = member, 19 = guest
			// 	}
			// 	// TODO: ask 7 how the fuck we get license info
			// 	// of guests when most people just put their
			// 	// own member number in this field instead of
			// 	// guest's ID information
			// 	if field.ID == "92" {
			// 		value = field.Value // this could in theory be LIC for guest...
			// 	}
			// }

			member, _ := makeNeonRequest("GET", fmt.Sprintf("https://api.neoncrm.com/v2/accounts/%v", element.AccountID), nil)
			var memberJson Account
			json.NewDecoder(member.Body).Decode(&memberJson)

			idFound := false
			for _, field := range memberJson.IndividualAccount.AccountCustomFields {
				if field.ID == "51" {
					idFound = true
					eventAttendeesMap[field.Value] = element.AccountID
					break
				}
			}
			if !idFound {
				// This is likely because the member is
				// actually a guest of another member?
				// Will need to acquire more context
				fmt.Printf("!!! ID NUMBER NOT FOUND !!! (account ID: %v)\n", element.AccountID)
				for _, field := range element.AttendeeCustomFields {
					fmt.Printf("%#v    ", field)
					if field.OptionValues != nil {
						fmt.Printf("%#v\n", *field.OptionValues)
					} else {
						fmt.Printf("\n")
					}
				}
			}
		}
		if msg.Pagination.CurrentPage == msg.Pagination.TotalPages {
			break
		} else {
			u, _ = url.Parse(fmt.Sprintf("https://api.neoncrm.com/v2/events/%v/attendees", eventId))
			q = u.Query()
			q.Set("currentPage", fmt.Sprint((msg.Pagination.CurrentPage + 1)))
			u.RawQuery = q.Encode()

			resp, _ = makeNeonRequest("GET", u.String(), nil)

			json.NewDecoder(resp.Body).Decode(&msg)
		}
	}

	// TODO: confirm why this sometimes doesn't
	// match given totalResults value (even
	// when missing IDs are added in)
	fmt.Printf("Total number of processed attendees: %v\n", len(eventAttendeesMap))

	body, _ := io.ReadAll(resp.Body)

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func verifyRegistration(w http.ResponseWriter, r *http.Request) {
	// TODO: business process/Neon query?
	// When do we know license needs updating
	// or waiver needs resigning?

	LIC := r.FormValue("LIC")
	DOB := r.FormValue("DOB")
	expiry := r.FormValue("expiry")
	origin := r.FormValue("origin")
	fmt.Printf("Received LIC: %v\n", LIC)
	fmt.Printf("Received DOB: %v\n", DOB)
	fmt.Printf("Received expiry: %v\n", expiry)
	fmt.Printf("Received origin: %v\n", origin)
	fmt.Println(eventAttendeesMap)

	// TODO: move the make() call of this map
	// to the updateAttendees endpoint, and
	// check to see if this map is nil prior
	// to checking for the LIC key (and return
	// an HTTP 503 or something)
	_, exists := eventAttendeesMap[LIC]
	fmt.Println(exists)

	if exists {
		// TODO: send PUT request to
		// set markedAttended as true
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
