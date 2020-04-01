package main

import (
	"flag"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"syscall"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/log"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	exampleBaseHref = "https://mycompany.teamwork.com"
	exampleUsername = "joe.bloggs@mycompany.com"
)

// clocks in from my last clockin date for every weekday through to today
// go run main.go
func main() {
	log := log.New()
	defer Recover(log.ErrorOn)
	log.Info("Starting")

	fs := flag.NewFlagSet("wtf_tw_clockin_catchup", flag.ExitOnError)
	var baseHref string
	fs.StringVar(&baseHref, "b", "", Sprintf("baseHref e.g. %s", exampleBaseHref))
	var username string
	fs.StringVar(&username, "u", "", Sprintf("username e.g. %s", exampleUsername))
	var hour int
	fs.IntVar(&hour, "h", 8, "hour e.g. 8")
	var minutes int
	fs.IntVar(&minutes, "m", 30, "minutes e.g. 30")
	var randomMinutes int64
	fs.Int64Var(&randomMinutes, "r", 0, "randomMinutes to vary start and end times e.g. 15")
	PanicOn(fs.Parse(os.Args[1:]))
	PanicIf(baseHref == "", "please enter a baseHref e.g. -b=%s", exampleBaseHref)
	PanicIf(username == "", "please enter a username e.g. -u=%s", exampleUsername)
	PanicIf(randomMinutes < 0 || randomMinutes > 60, "please enter a randomMinutes between 0 and 59 e.g. -r=15")

	Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	PanicOn(err)
	pwd := string(bytePassword)
	Println()

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

	for clockout.Before(Now()) {
		if !(clockin.Weekday() == time.Saturday || clockin.Weekday() == time.Sunday) {
			clockinToUse := clockin
			clockoutToUse := clockout
			if randomMinutes > 0 {
				posNegVar := (rand.Int63n(1) * -1) + 1
				clockinToUse = clockinToUse.Add(time.Duration(rand.Int63n(randomMinutes)*posNegVar) * time.Minute)
				clockoutToUse = clockoutToUse.Add(time.Duration(rand.Int63n(randomMinutes)*posNegVar) * time.Minute)
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
	PanicOn(err)
	req.SetBasicAuth(username, pwd)
	resp, err := http.DefaultClient.Do(req)
	PanicOn(err)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	PanicOn(err)
	PanicIf(resp.StatusCode > 299, "error making request, status code: %d, status: %s body: %s", resp.StatusCode, resp.Status, string(respBody))
	return json.MustFromBytes(respBody)
}
