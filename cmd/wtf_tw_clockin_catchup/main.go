package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"syscall"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	exampleBaseHref = "https://mycompany.teamwork.com"
	exampleUsername = "joe.bloggs@mycompany.com"
)

// clocks in from my last clockin date for every weekday through to today
// go run main.go
func main() {
	fmt.Println("Starting")

	fs := flag.NewFlagSet("wtf_tw_clockin_catchup", flag.ExitOnError)
	var baseHref string
	fs.StringVar(&baseHref, "b", "", fmt.Sprintf("baseHref e.g. %s", exampleBaseHref))
	var username string
	fs.StringVar(&username, "u", "", fmt.Sprintf("username e.g. %s", exampleUsername))
	var hour int
	fs.IntVar(&hour, "h", 8, "hour e.g. 8")
	var minutes int
	fs.IntVar(&minutes, "m", 30, "minutes e.g. 30")
	log := NewLog(nil, false, true)
	ExitOn(fs.Parse(os.Args[1:]))
	ExitIf(baseHref == "", "please enter a baseHref e.g. -b=%s", exampleBaseHref)
	ExitIf(username == "", "please enter a username e.g. -u=%s", exampleUsername)

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	ExitOn(err)
	pwd := string(bytePassword)
	fmt.Println()

	// get my user id
	myID := mustDoReq(http.MethodGet, baseHref, "/me.json", username, pwd, nil).MustInt64("person", "id")
	log.Info("myID: %d", myID)

	// get last clockin
	now := Now()
	clockin := time.Date(now.Year(), now.Month(), now.Day(), hour, minutes, 0, 0, time.Local)
	resp := mustDoReq(http.MethodGet, baseHref, "/me/clockins.json?pageSize=1", username, pwd, nil)
	if resp.Exists("clockIns", 0, "clockInDatetime") {
		clockin = resp.MustTime("clockIns", 0, "clockInDatetime").Add(time.Hour * 24)
		clockin = time.Date(clockin.Year(), clockin.Month(), clockin.Day(), hour, minutes, 0, 0, time.Local)
	}
	log.Info("starting at clockin: %s", clockin)
	clockout := clockin.Add((8 * time.Hour) + (30 * time.Minute))

	for clockout.Before(time.Now()) {
		if !(clockin.Weekday() == time.Saturday || clockin.Weekday() == time.Sunday) {
			mustDoReq(http.MethodPost, baseHref, "/clockin.json", username, pwd,
				json.MustFromString(`{"clockIn":{}}`).
					MustSet("clockIn", "userId", myID).
					MustSet("clockIn", "clockInDatetime", clockin.Format("20060102150405")).
					MustSet("clockIn", "clockOutDatetime", clockout.Format("20060102150405")).
					MustToReader())
			log.Info("clocking in: clockin: %s clockout: %s", clockin, clockout)
		}
		clockin = clockin.Add(time.Hour * 24)
		clockout = clockout.Add(time.Hour * 24)
	}
	log.Info("Finished")
}

func mustDoReq(method, baseHref, path, username, pwd string, body io.Reader) *json.Json {
	req, err := http.NewRequest(method, baseHref+path, body)
	ExitOn(err)
	req.SetBasicAuth(username, pwd)
	resp, err := http.DefaultClient.Do(req)
	ExitOn(err)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	ExitOn(err)
	ExitIf(resp.StatusCode > 299, "error making request, status code: %d, status: %s body: %s", resp.StatusCode, resp.Status, string(respBody))
	return json.MustFromBytes(respBody)
}
