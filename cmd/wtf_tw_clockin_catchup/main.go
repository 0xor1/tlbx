package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
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
	var randomMinutes int64
	fs.Int64Var(&randomMinutes, "r", 0, "randomMinutes to vary start and end times e.g. 15")
	log := GetLog()
	ExitOn(fs.Parse(os.Args[1:]))
	ExitIf(baseHref == "", "please enter a baseHref e.g. -b=%s", exampleBaseHref)
	ExitIf(username == "", "please enter a username e.g. -u=%s", exampleUsername)
	ExitIf(randomMinutes < 0 || randomMinutes > 60, "please enter a randomMinutes between 0 and 59 e.g. -r=15")

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	ExitOn(err)
	pwd := string(bytePassword)
	fmt.Println()

	rand.Seed(NowUnixMilli())

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
			clockinToUse := clockin
			clockoutToUse := clockout
			if randomMinutes > 0 {
				clockinToUse = clockinToUse.Add(time.Duration(rand.Int63n(randomMinutes)) * time.Minute)
				clockoutToUse = clockoutToUse.Add(time.Duration(rand.Int63n(randomMinutes)) * time.Minute)
			}
			mustDoReq(http.MethodPost, baseHref, "/clockin.json", username, pwd,
				json.MustFromString(`{"clockIn":{}}`).
					MustSet("clockIn", "userId", myID).
					MustSet("clockIn", "clockInDatetime", clockinToUse.Format("20060102150405")).
					MustSet("clockIn", "clockOutDatetime", clockoutToUse.Format("20060102150405")).
					MustToReader())
			log.Info("clocking in: clockin: %s clockout: %s", clockinToUse, clockoutToUse)
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
