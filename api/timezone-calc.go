package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

const (
	hSignature = "X-Slack-Signature"
	hTimestamp = "X-Slack-Request-Timestamp"
)

var slackSigningSecret string

func init() {
	// set SLACK_SIGNING_SECRET=<Signing Secret of your Slack App> in windows cli
	slackSigningSecret = os.Getenv("SLACK_SIGNING_SECRET")
	fmt.Println(slackSigningSecret)
}

func checkTime(t time.Time) bool {

	day := t.Weekday()
	hours := t.Hour()
	fmt.Println("day: ", day, " hours: ", hours)

	return (day >= 1 &&
		day <= 5 &&
		hours >= 9 &&
		hours < 18)
}

// TimeZoneCurrentTime is Translation function
func TimeZoneCurrentTime(w http.ResponseWriter, r *http.Request) {

	// Verify Slack Request with Signing Secret, and Timeout check
	if ok := verifySlackSignature(r, []byte(slackSigningSecret)); ok == false {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed on VerifyRequest()")
		return
	}

	// Parsing Slack Slash Command
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to parse Slack Slash Command: %v", err)
		return
	}

	// TZDN: TimeZone database name
	timezonDatabaseName := s.Text

	fmt.Println("TimeZone: ", timezonDatabaseName)

	// get loc for get timezone current time
	loc, err := time.LoadLocation(timezonDatabaseName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to time.LoadLocation: %v", err)
		return
	}
	timezoneCurrentTime := time.Now().In(loc)
	isBusinessHour := checkTime(timezoneCurrentTime)
	fmt.Println("current Time: ", timezoneCurrentTime, " and Business Hour? ", isBusinessHour)

	// slactPost := fmt.Sprintf("`source`: %s\n`target`: %s\n", srcText, tgtText)
	// params := &slack.Msg{
	// 	Type: "mrkdwn",
	// 	Text: slactPost,
	// }

	params := &slack.Msg{
		Type: "mrkdwn",
		Text: timezoneCurrentTime.String(),
	}

	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// TimeZoneCurrentTime is Translation function
func TimeZoneCurrentTimeNewYork(w http.ResponseWriter, r *http.Request) {

	// Verify Slack Request with Signing Secret, and Timeout check
	if ok := verifySlackSignature(r, []byte(slackSigningSecret)); ok == false {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed on VerifyRequest()")
		return
	}

	// TZDN: TimeZone database name
	newYorkTimezonDatabaseName := "America/New_York" // https://www.wikiwand.com/en/List_of_tz_database_time_zones

	// get loc for get timezone current time
	loc, _ := time.LoadLocation(newYorkTimezonDatabaseName)
	timezoneCurrentTime := time.Now().In(loc)

	// slactPost := fmt.Sprintf("`source`: %s\n`target`: %s\n", srcText, tgtText)
	// params := &slack.Msg{
	// 	Type: "mrkdwn",
	// 	Text: slactPost,
	// }

	params := &slack.Msg{
		Type: "mrkdwn",
		Text: timezoneCurrentTime.String(),
	}

	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, receivedMAC string, slackSigningSecret []byte) bool {
	mac := hmac.New(sha256.New, slackSigningSecret)
	if _, err := mac.Write([]byte(message)); err != nil {
		log.Printf("mac.Write(%v) failed\n", message)
		return false
	}
	calculatedMAC := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(calculatedMAC), []byte(receivedMAC))
}

// VerifySlackSignature verifies the request is coming from Slack
// Read https://api.slack.com/docs/verifying-requests-from-slack
func verifySlackSignature(r *http.Request, slackSigningSecret []byte) bool {
	if r.Body == nil {
		return false
	}

	// do not consume req.body
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// prepare message for signing
	timestamp := r.Header.Get(hTimestamp)
	slackSignature := r.Header.Get(hSignature)
	message := "v0:" + timestamp + ":" + string(bodyBytes)

	// Timeout check
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Printf("failed strconv.ParseInt%v\n", err)
		return false
	}

	tSince := time.Since(time.Unix(ts, 0))
	diff := time.Duration(abs64(int64(tSince)))
	if diff > 5*time.Minute {
		log.Println("timed out")
		return false
	}

	// Not timeouted, then check Mac
	return checkMAC(message, slackSignature, slackSigningSecret)
}

func abs64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}
