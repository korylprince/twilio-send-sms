```
$ twilio-send-sms -h
twilio-send-sms [flags]
twilio-send-sms sends an SMS message using the Twilio API. Parameters can be set using flags or environment variables, and the message body should be passed to stdin.
Accepted flags are:
  -from string
    	from phone number (SMS_FROM)
  -shortener string
    	URL shortener endpoint. Should contain %s where URL will be passed (SMS_SHORTENER)
  -sid string
    	Twilio Account SID (SMS_ACCOUNTSID)
  -to string
    	to phone number (SMS_TO)
  -token string
    	Twilio Auth Token (SMS_AUTHTOKEN)
  -url string
    	URL to shorten and send (SMS_URL)
```
