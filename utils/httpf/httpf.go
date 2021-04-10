package httpf

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Errors
var (
	ErrReadBody          = errors.New("exchange internal error happened")
	ErrUnknownHTTPMethod = errors.New("unknown http method")
	ErrParseURL          = errors.New("can't parse url")
	ErrQueryNil          = errors.New("query is nil")
	ErrCreateRequest     = errors.New("can't create request")
	ErrSetCertificate    = errors.New("set certificate error")
)

// Get request
func Get(
	URL string,
	query url.Values,
	header http.Header,
) (body []byte, err error) {

	if query == nil {
		query = url.Values{}
	}

	// set transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{},
	}

	// client settings
	client := &http.Client{
		Timeout:   time.Second * 3,
		Transport: transport,
	}
	// new GET request
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, ErrCreateRequest
	}

	// add query and headers
	if header != nil {
		req.Header = header
	}
	// add query
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v[0])
	}
	req.URL.RawQuery = q.Encode()
	//log.Println(req.URL.String())
	// DO
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrReadBody
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error: " + string(body))
	}
	return body, nil
}

// Post request
func Post(
	URL string,
	header http.Header,
	query url.Values,
	data string,
) (body []byte, err error) {

	if query == nil {
		return nil, ErrQueryNil
	}

	// set transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// client settings
	client := &http.Client{
		Timeout:   time.Second * 3,
		Transport: transport,
	}

	// new POST request
	rbody := bytes.NewBufferString(data)
	req, err := http.NewRequest(http.MethodPost, URL, rbody)
	if err != nil {
		return nil, ErrCreateRequest
	}
	// add query and headers
	req.Header = header
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v[0])
	}
	req.URL.RawQuery = q.Encode()
	//log.Println(req.URL.String())
	// DO
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrReadBody
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated {
		return nil, errors.New("error: " + string(body))
	}
	return body, nil
}

// Delete request
func Delete(
	URL string,
	header http.Header,
	query url.Values,
	data string,
) (body []byte, err error) {

	if query == nil {
		return nil, ErrQueryNil
	}

	// set transport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// client settings
	client := &http.Client{
		Timeout:   time.Second * 3,
		Transport: transport,
	}

	// new POST request
	rbody := bytes.NewBufferString(data)
	req, err := http.NewRequest(http.MethodDelete, URL, rbody)
	if err != nil {
		return nil, ErrCreateRequest
	}
	// add query and headers
	req.Header = header
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v[0])
	}
	req.URL.RawQuery = q.Encode()
	//log.Println(req.URL.String())
	// DO
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrReadBody
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated {
		return nil, errors.New("error: " + string(body))
	}
	return body, nil
}
