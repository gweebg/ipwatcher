package watcher

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gweebg/ipwatcher/internal/config"
)

var fetchLogger = logger.With().Str("service", "fetcher").Logger()

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
			fetchLogger.Error().Err(err).Str("source_name", source.Name).Msgf("source does not have any 'IP%v' url specified, skipping", version)
			continue
		}

		response, err := sendRequest(url)
		if err != nil {
			fetchLogger.Error().Err(err).Str("source_name", source.Name).Msg("failed to send request to source")
			continue
		}

		address = parseResponse(response, source)

		valid := net.ParseIP(address)
		if valid == nil {
			fetchLogger.Error().Err(err).Str("source_name", source.Name).Msgf("source did not return a valid IP address: '%v', skipping", address)
			continue
		}

		fromSource = url
		fetchLogger.Debug().Str("source", url).Msgf("valid address from source '%v'", source.Name)

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
		fetchLogger.Warn().Msgf("'%v' returned %d", source.Name, response.StatusCode)
		return ""
	}

	// handle the different Content-Types
	contentType := response.Header.Get("Content-Type")

	// todo: abstract these if's for easier plug and play
	if source.Type == "text" && strings.Contains(contentType, "text/plain") {

		var address bytes.Buffer
		_, err := io.Copy(&address, response.Body)
		if err != nil {
			fetchLogger.Error().Err(err).Str("source_name", source.Name).Msg("error reading response body")
			return ""
		}

		return address.String()
	}

	if source.Type == "json" && strings.Contains(contentType, "application/json") {

		var responseBody map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		if err != nil {
			fetchLogger.Error().Err(err).Str("source_name", source.Name).Msg("error decoding JSON response")
			return ""
		}

		address, ok := responseBody[*source.Field]
		if !ok {
			fetchLogger.Error().Str("source_name", source.Name).Msgf("expected field '%v' to be present on response", *source.Field)
			return ""
		}

		return address.(string)
	}

	fetchLogger.Error().Str("source_name", source.Name).Msgf("content type between response and config mismatch, expected '%s' but got '%s'", source.Type, contentType)
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
