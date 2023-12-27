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

type AttendanceStatus struct {
	Name       string `json:"name"`
	Capacity   int    `json:"capacity"`
	Registered int    `json:"registered"`
	database   map[string]string
	checkedIn  map[string]struct{}
}

type ServerStatus struct {
	Response string `json:"response,omitempty"`
	// TODO: do I really need to omitempty these?
	// ETA: if I change this, I'll need to
	// change the logic on the frontend to
	// not look for undefined when receiving
	// server status response
	ListedEvents    []Event                      `json:"listedEvents,omitempty"`
	EventAttendance map[string]*AttendanceStatus `json:"eventAttendance,omitempty"`
}

var client = &http.Client{}

var eventAttendanceDatabase map[string]*AttendanceStatus
var currentlyListedEvents []Event

func main() {
	eventAttendanceDatabase = make(map[string]*AttendanceStatus)
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

func getServerStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(ServerStatus{
		EventAttendance: eventAttendanceDatabase,
		ListedEvents:    currentlyListedEvents,
	})
	w.Write(resp)
}

func refreshEvent(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse("https://api.neoncrm.com/v2/events")
	q := u.Query()
	// Subtracting a day just to fuzz the numbers
	// in case for some reason it doesn't want to
	// give us the events that are occurring today
	q.Set("startDateAfter", time.Now().Add(-time.Hour*24).Format(time.DateOnly))
	q.Set("startDateBefore", time.Now().Add(time.Hour*24*7).Format(time.DateOnly)) // TODO: these values should not be magic
	u.RawQuery = q.Encode()

	resp, err := makeNeonRequest("GET", u.String(), nil)

	var msg EventRequest
	var decResp []byte

	if err != nil {
		// do something to fix error
	} else if resp.StatusCode != 200 {
		// do something to note non-200 response
	} else {
		json.NewDecoder(resp.Body).Decode(&msg)
		currentlyListedEvents = msg.Events

		decResp, _ = json.Marshal(ServerStatus{
			EventAttendance: eventAttendanceDatabase,
			ListedEvents:    currentlyListedEvents,
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(decResp)
	}
}

func addEvent(w http.ResponseWriter, r *http.Request) {
	// TODO: sanitize input of attendees by
	// upcasing ID values before putting
	// into attendee map

	// TODO: check for markedAttended true
	// before adding to map? to make sure
	// that if we have to restart from crash
	// we have up-to-date attendance info

	eventId := r.FormValue("eventId")
	eventName := r.FormValue("eventName")

	u, _ := url.Parse(fmt.Sprintf("https://api.neoncrm.com/v2/events/%v/attendees", eventId))
	q := u.Query()
	u.RawQuery = q.Encode()

	resp, _ := makeNeonRequest("GET", u.String(), nil)

	var msg EventAttendees
	eventAttendeesMap := make(map[string]string)

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
	fmt.Println(eventAttendeesMap)

	eventAttendanceDatabase[eventId] = &AttendanceStatus{
		Name:       eventName,
		Capacity:   len(eventAttendeesMap),
		Registered: 0,
		database:   eventAttendeesMap,
		checkedIn:  make(map[string]struct{}),
	}

	eventResponse, _ := json.Marshal(ServerStatus{
		EventAttendance: eventAttendanceDatabase,
		ListedEvents:    currentlyListedEvents,
	})

	w.WriteHeader(http.StatusCreated)
	w.Write(eventResponse)
}

func verifyRegistration(w http.ResponseWriter, r *http.Request) {
	// TODO: business process/Neon query?
	// When do we know license needs updating
	// or waiver needs resigning?

	// TODO: check to see if len(eventMap)
	// is 0 prior to checking for the LIC key
	// (and return an HTTP 503 or something)

	// TODO: log registrations!

	// Gather all form values
	// from received request
	licJson := r.FormValue("LIC")
	dobJson := r.FormValue("DOB")
	expiryJson := r.FormValue("expiry")
	originJson := r.FormValue("origin")

	// Format of expiry and DOB on
	// license differs depending
	// on country of origin
	var timeLayout string
	if originJson == "USA" {
		timeLayout = "01022006"
	} else {
		timeLayout = "20060102"
	}

	dobParsed, _ := time.Parse(timeLayout, dobJson)
	yearsSinceBirthYear := time.Now().Year() - dobParsed.Year()

	// TODO: factor into own function
	isUnderage := false
	if yearsSinceBirthYear < 18 {
		isUnderage = true
	} else if yearsSinceBirthYear == 18 {
		monthsSinceBirthMonth := time.Now().Month() - dobParsed.Month()
		if monthsSinceBirthMonth < 0 {
			isUnderage = true
		} else if monthsSinceBirthMonth == 0 {
			daysSinceBirthDay := time.Now().Day() - dobParsed.Day()
			if daysSinceBirthDay < 0 {
				isUnderage = true
			}
		}
	}
	if isUnderage {
		w.WriteHeader(http.StatusForbidden)
		resp, _ := json.Marshal(ServerStatus{
			Response:        "underage",
			EventAttendance: eventAttendanceDatabase,
			ListedEvents:    currentlyListedEvents,
		})
		w.Write(resp)
		return
	}

	expiryParsed, _ := time.Parse(timeLayout, expiryJson)
	untilExpiry := time.Until(expiryParsed)

	if untilExpiry < 0 {
		w.WriteHeader(http.StatusForbidden)
		resp, _ := json.Marshal(ServerStatus{
			Response:        "expired",
			EventAttendance: eventAttendanceDatabase,
			ListedEvents:    currentlyListedEvents,
		})
		w.Write(resp)
		return
	}

	// If we're here, the cardholder is
	// of age and the ID is unexpired

	var exists bool
	var event string
	for id, eventStatus := range eventAttendanceDatabase {
		_, inMap := eventStatus.database[licJson]
		if inMap {
			exists = true
			event = id
			eventStatus.checkedIn[licJson] = struct{}{}
			eventStatus.Registered = len(eventAttendanceDatabase[event].checkedIn)
		}
	}

	regStat := ServerStatus{
		EventAttendance: eventAttendanceDatabase,
		ListedEvents:    currentlyListedEvents,
	}

	if exists {
		// TODO: send PUT request to
		// set markedAttended as true
		regStat.Response = event // maybe needs to be a pointer?
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	response, _ := json.Marshal(regStat)
	w.Write(response)
}
