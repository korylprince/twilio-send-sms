package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//APIURL is the Twilio API Messages endpoint
const APIURL = "https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json"

var fromNumber = flag.String("from", os.Getenv("SMS_FROM"), "from phone number (SMS_FROM)")
var toNumber = flag.String("to", os.Getenv("SMS_TO"), "to phone number (SMS_TO)")
var accountSID = flag.String("sid", os.Getenv("SMS_ACCOUNTSID"), "Twilio Account SID (SMS_ACCOUNTSID)")
var authToken = flag.String("token", os.Getenv("SMS_AUTHTOKEN"), "Twilio Auth Token (SMS_AUTHTOKEN)")
var shortenURL = flag.String("shortener", os.Getenv("SMS_SHORTENER"), "URL shortener endpoint. Should contain %s where URL will be passed (SMS_SHORTENER)")
var smsURL = flag.String("url", os.Getenv("SMS_URL"), "URL to shorten and send (SMS_URL)")

func init() {
	flag.Usage = func() {
		fmt.Println(os.Args[0], "[flags]")
		fmt.Println("twilio-send-sms sends an SMS message using the Twilio API. Parameters can be set using flags or environment variables, and the message body should be passed to stdin.\nAccepted flags are:")
		flag.PrintDefaults()
	}
}

//TrimBody trims the given message body (including URL) to 160 characters
func TrimBody(body, url string) string {
	if len(url) > 0 {
		url = "\n" + url
	}
	if len(body+url) <= 160 {
		return body + url
	}
	return body[:157-len(url)] + "..." + url
}

//Shorten returns a shortened version of the given URL or an error if one occurs
func Shorten(shortenURL, url string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(shortenURL, url))
	if err != nil {
		return "", fmt.Errorf("Unable to get shortened URL: %v", err)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", fmt.Errorf("Unable read response body: %v", err)
	}

	short := strings.TrimSpace(buf.String())
	if short == "" {
		return "", fmt.Errorf("Unable read response body: %v", err)
	}

	return short, nil
}

//Send sends the message with given details or returns and error if one occurred
func Send(from, to, body, accountSID, authToken string) error {
	//truncate body to 160 characters
	if len(body) > 160 {
		body = body[:160]
	}

	//set body
	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	data.Set("Body", body)
	reader := strings.NewReader(data.Encode())

	//create request
	req, err := http.NewRequest("POST", fmt.Sprintf(APIURL, accountSID), reader)
	if err != nil {
		return fmt.Errorf("Unable to create request: %v", err)
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	//send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to complete request: %v", err)
	}

	//request successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	//request failed
	var respBody map[string]interface{}

	d := json.NewDecoder(resp.Body)
	err = d.Decode(&respBody)
	if err != nil {
		return fmt.Errorf("Unable to decode response body: %v", err)
	}

	return fmt.Errorf("HTTP error %d, %s %s", resp.StatusCode, respBody["message"], respBody["more_info"])

}

func main() {
	//parse flags
	flag.Parse()

	if *fromNumber == "" {
		flag.Usage()
		fmt.Println("\nError: -from flag or SMS_FROM environment variable must be supplied")
		os.Exit(1)
	}

	if *toNumber == "" {
		flag.Usage()
		fmt.Println("\nError: -to flag or SMS_TO environment variable must be supplied")
		os.Exit(1)
	}

	if *accountSID == "" {
		flag.Usage()
		fmt.Println("\nError: -sid flag or SMS_ACCOUNTSID environment variable must be supplied")
		os.Exit(1)
	}

	if *authToken == "" {
		flag.Usage()
		fmt.Println("\nError: -token flag or SMS_AUTHTOKEN environment variable must be supplied")
		os.Exit(1)
	}

	//read message from stdin
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, os.Stdin); err != nil {
		fmt.Println("Unable to read from stdin:", err)
		os.Exit(1)
	}

	var (
		short string
		err   error
	)

	if *smsURL != "" {
		if *shortenURL == "" {
			flag.Usage()
			fmt.Println("\nIf -url (SMS_URL) is given, -shortener (SMS_SHORTENER) must also be given")
			os.Exit(1)
		}

		short, err = Shorten(*shortenURL, *smsURL)
		if err != nil {
			fmt.Println("Unable to get shortened URL:", err)
			os.Exit(1)
		}
	}

	err = Send(*fromNumber, *toNumber, TrimBody(buf.String(), short), *accountSID, *authToken)
	if err != nil {
		fmt.Println("Unable to send SMS:", err)
		os.Exit(1)
	}
}
