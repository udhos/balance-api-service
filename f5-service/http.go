package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func tlsConfig() *tls.Config {
	return &tls.Config{
		//CipherSuites:             []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA},
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
		//MaxVersion:               tls.VersionTLS11,
		//MinVersion:               tls.VersionTLS11,
	}
}

func httpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig:    tlsConfig(),
		DisableCompression: true,
		DisableKeepAlives:  true,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
}

func httpPostString(url string, contentType string, s string) ([]byte, error) {
	c := httpClient()
	return clientPost(c, url, contentType, bytes.NewBufferString(s))
}

func httpGet(url string) ([]byte, error) {
	c := httpClient()
	return clientGet(c, url)
}

func clientPost(c *http.Client, url string, contentType string, r io.Reader) ([]byte, error) {

	resp, errPost := c.Post(url, contentType, r)
	if errPost != nil {
		return nil, errPost
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("httpPost: bad status: %d", resp.StatusCode)
	}

	body, errBody := ioutil.ReadAll(resp.Body)
	if errBody != nil {
		return nil, fmt.Errorf("httpPost: read: url=%v: %v", url, errBody)
	}

	return body, errBody
}

func clientGet(c *http.Client, url string) ([]byte, error) {
	resp, errGet := c.Get(url)
	if errGet != nil {
		return nil, fmt.Errorf("httpGet: get url=%v: %v", url, errGet)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("httpGet: bad status: %d", resp.StatusCode)
	}

	info, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return nil, fmt.Errorf("httpGet: read all: url=%v: %v", url, errRead)
	}

	return info, errRead
}

func writeStr(caller string, w http.ResponseWriter, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		log.Printf("%s writeStr: %v", caller, err)
	}
}

func sendBadRequest(label, reason string, w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s: method=%s url=%s from=%s - bad request: %s", label, r.Method, r.URL.Path, r.RemoteAddr, reason)
	log.Print(msg)

	http.Error(w, reason, http.StatusBadRequest) // 400
}

func sendNotSupported(label string, w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s: method=%s url=%s from=%s - method not supported", label, r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	w.Header().Set("Allow", "POST") // required by 405 error

	http.Error(w, r.Method+" method not supported", http.StatusMethodNotAllowed) // 405
}

func sendNotFound(label string, w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s: method=%s url=%s from=%s - PATH NOT FOUND", label, r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	notFound := fmt.Sprintf("path not found: [%s]", r.URL.Path)

	w.WriteHeader(http.StatusNotFound)

	io.WriteString(w, notFound+"\n")
}

func sendNotImplemented(label string, w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s: method=%s url=%s from=%s - NOT IMPLEMENTED", label, r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	http.Error(w, label+" not implemented", http.StatusNotImplemented) // 501
}
