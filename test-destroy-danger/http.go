package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
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

func httpGet(url, contentType, user, pass string) ([]byte, error) {
	c := httpClient()
	return clientMethod(c, "GET", url, contentType, user, pass, nil)
}

func httpPut(url, contentType, user, pass, s string) ([]byte, error) {
	c := httpClient()
	return clientMethod(c, "PUT", url, contentType, user, pass, bytes.NewBufferString(s))
}

func clientMethod(c *http.Client, method, url, contentType, user, pass string, body io.Reader) ([]byte, error) {

	req, errNew := http.NewRequest(method, url, body)
	if errNew != nil {
		return nil, errNew
	}
	req.Header.Set("Content-Type", contentType)

	req.SetBasicAuth(user, pass)

	resp, errDel := c.Do(req)
	if errDel != nil {
		return nil, errDel
	}

	defer resp.Body.Close()

	info, errRead := ioutil.ReadAll(resp.Body)
	if errRead != nil {
		return info, fmt.Errorf("http method=%s: read all: url=%v: %v", method, url, errRead)
	}

	if resp.StatusCode != 200 {
		return info, fmt.Errorf("http method=%s: bad status: %d", method, resp.StatusCode)
	}

	return info, nil
}
