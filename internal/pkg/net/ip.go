package net

import (
	"io"
	"net/http"
	"regexp"
	"time"
)

const (
	regexIPv4     = `(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)`
	ipURL         = "http://cip.cc"
	timeout       = 10 // seconds
	retry         = 3
	retryInterval = 10 // seconds
)

// TODO: support multiple IP urls with channel, make this robust
func GetIP() (string, error) {
	var retErr error
	for i := 0; i < retry; i++ {
		ip, err := getIP()
		if err == nil {
			return ip, nil
		}

		retErr = err
		time.Sleep(retryInterval * time.Second)
		continue
	}

	return "", retErr
}

func getIP() (string, error) {
	req, err := http.NewRequest("GET", ipURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 7.0)")
	client := &http.Client{
		Timeout: timeout * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return regexp.MustCompile(regexIPv4).FindString(string(body)), nil
}
