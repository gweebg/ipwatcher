package watcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gweebg/ipwatcher/internal/config"
)

func RequestAddress(version string) (string, string, error) {

	conf := config.GetConfig()
	sources := conf.Get("sources").([]config.Source)

	address := ""
	fromSource := ""

	const forceSource = "service.force_source"
	for _, source := range sources {

		// if 'force_source' is set, and it is different from the current source, we skip it
		if conf.IsSet(forceSource) && conf.Get(forceSource) != source.Name {
			continue
		}

		url, err := source.Url.GetUrl(version)
		if err != nil {
			log.Printf("source '%v' does not have any 'IP%v' url specified, skipping\n", source.Name, version)
			continue
		}

		response, err := sendRequest(url)
		if err != nil {
			log.Printf("failed to send request to source '%v', skipping\n", source.Name)
			continue
		}

		address = parseResponse(response, source)

		valid := net.ParseIP(address)
		if valid == nil {
			log.Printf("source '%v' did not return a valid IP address: '%v', skipping\n", source.Name, address)
			continue
		}

		fromSource = url
		log.Printf("valid address from source '%v'\n", source.Name)
		break
	}

	// if address is still empty after querying the urls then, user needs to try others
	if address == "" {
		return address, "", errors.New(
			"none of the specified sources returned a valid address or 'force_source' name mismatch",
		)
	}

	return address, fromSource, nil
}

func parseResponse(response *http.Response, source config.Source) string {

	// check if http response status code is 'positive' (200<=status<300)
	if !(response.StatusCode >= 200 && response.StatusCode < 300) {
		log.Printf("'%v' returned %d\n", source.Name, response.StatusCode)
		return ""
	}

	// handle the different Content-Types
	contentType := response.Header.Get("Content-Type")

	// todo: abstract these if's for easier plug and play
	if source.Type == "text" && strings.Contains(contentType, "text/plain") {

		var address bytes.Buffer
		_, err := io.Copy(&address, response.Body)
		if err != nil {
			log.Printf("error reading response body from source '%v'", source.Name)
			return ""
		}

		return address.String()
	}

	if source.Type == "json" && strings.Contains(contentType, "application/json") {

		var responseBody map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		if err != nil {
			log.Printf("error decoding JSON response from source '%v'", source.Name)
			return ""
		}

		address, ok := responseBody[*source.Field]
		if !ok {
			log.Printf("expected field '%v' present on response:\n\t%v", *source.Field, responseBody)
			return ""
		}

		return address.(string)
	}

	log.Printf("content type between response and config mismatch, expected '%s' but got '%s'\n", source.Type, contentType)
	return ""
}

func sendRequest(url string) (*http.Response, error) {

	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	return client.Do(req)
}
