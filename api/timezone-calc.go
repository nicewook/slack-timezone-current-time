package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	//fmt.Println(slackSigningSecret)
}

func checkTime(t time.Time) bool {

	day := t.Weekday()
	hours := t.Hour()

	return (day >= 1 &&
		day <= 5 &&
		hours >= 9 &&
		hours < 18)
}

func makeResponse(timezonDatabaseName string) string {
	loc, err := time.LoadLocation(timezonDatabaseName)
	if err != nil {
		// w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to time.LoadLocation: %v", err)
		return ""
	}
	timezoneCurrentTime := time.Now().In(loc)
	isBusinessHour := checkTime(timezoneCurrentTime)

	var bHourFormat string
	if isBusinessHour {
		bHourFormat = "`it's Business Hour` in *" + timezonDatabaseName + "* :grinning::grinning::grinning::grinning::grinning:"
	} else {
		bHourFormat = "`it's Not Business Hour` in *" + timezonDatabaseName + "* :angry::angry::angry::angry::angry:"
	}

	// Prepare response
	tzFormat := timezoneCurrentTime.Format("2006-01-02 15:04:05 Monday MST")
	responseText := ">" + bHourFormat + "\n>" + tzFormat + "\n"
	return responseText
}

// TimeZoneCurrentTime handles /tz Slack Slash Command
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

	// Get all Matching TimeZone database name using regexp
	// and get Timezone Current Time
	var responseText string
	for _, tzdbName := range timezoneNameArray {
		matched, _ := regexp.MatchString(strings.ToLower(s.Text), strings.ToLower(tzdbName))
		if matched {
			resp := makeResponse(tzdbName)
			responseText += resp
		}
	}

	if responseText == "" {
		responseText = "No match for " + s.Text
	}

	params := &slack.Msg{
		Type: "mrkdwn",
		Text: responseText,
	}

	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// TimeZoneCurrentTimeNewYork handles /tzn Slack Slash Command
func TimeZoneCurrentTimeNewYork(w http.ResponseWriter, r *http.Request) {

	// Verify Slack Request with Signing Secret, and Timeout check
	if ok := verifySlackSignature(r, []byte(slackSigningSecret)); ok == false {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed on VerifyRequest()")
		return
	}

	responseText := makeResponse("America/New_York") // https://www.wikiwand.com/en/List_of_tz_database_time_zones
	params := &slack.Msg{
		Type: "mrkdwn",
		Text: responseText,
	}

	b, err := json.Marshal(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// TimeZoneCurrentTimeSeoul handles /tzs Slack Slash Command
func TimeZoneCurrentTimeSeoul(w http.ResponseWriter, r *http.Request) {

	// Verify Slack Request with Signing Secret, and Timeout check
	if ok := verifySlackSignature(r, []byte(slackSigningSecret)); ok == false {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("failed on VerifyRequest()")
		return
	}

	responseText := makeResponse("Asia/Seoul") // https://www.wikiwand.com/en/List_of_tz_database_time_zones
	params := &slack.Msg{
		Type: "mrkdwn",
		Text: responseText,
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
