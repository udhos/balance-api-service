package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
)

const (
	version = "0.0"
)

var (
	showPasswords bool
)

func main() {

	me := os.Args[0]

	log.Printf("%s %s runtime %s GOMAXPROCS=%d", me, version, runtime.Version(), runtime.GOMAXPROCS(0))

	var debug bool
	if os.Getenv("DEBUG") != "" {
		debug = true
	}
	log.Printf("debug=%v DEBUG=[%s]", debug, os.Getenv("DEBUG"))

	addr := os.Getenv("LISTEN")
	if addr == "" {
		addr = ":8080"
	}

	if os.Getenv("SHOW_PASSWORDS") != "" {
		showPasswords = true
	}
	log.Printf("showPasswords=%v SHOW_PASSWORDS=[%s]", showPasswords, os.Getenv("SHOW_PASSWORDS"))

	dry := true
	if os.Getenv("NO_DRY") != "" {
		dry = false
	}
	log.Printf("dry=%v NO_DRY=[%s]", dry, os.Getenv("NO_DRY"))

	key := os.Getenv("KEY")
	if key == "" {
		key = "key.pem"
	}
	cert := os.Getenv("CERT")
	if cert == "" {
		cert = "cert.pem"
	}
	log.Printf("TLS key=%s cert=%s -- change with env vars: KEY=[%s] CERT=[%s]", key, cert, os.Getenv("KEY"), os.Getenv("CERT"))

	tls := true
	if !fileExists(key) {
		log.Printf("TLS key file not found: %s - disabling TLS", key)
		tls = false
	}
	if !fileExists(cert) {
		log.Printf("TLS cert file not found: %s - disabling TLS", cert)
		tls = false
	}

	register("/", func(w http.ResponseWriter, r *http.Request) { handlerRoot(w, r, "/") })

	register("/v1/ff/node/", func(w http.ResponseWriter, r *http.Request) { handlerNodeF5(w, r, "/v1/ff/node/") })
	register("/v1/at2/node/", func(w http.ResponseWriter, r *http.Request) { handlerNodeA10v2(debug, dry, w, r, "/v1/at2/node/") })
	register("/v1/at2/healthcheck", func(w http.ResponseWriter, r *http.Request) { handlerNodeA10v2Health(w, r, "/v1/at2/healthcheck") })
	register("/v1/at2/healthcheck/", func(w http.ResponseWriter, r *http.Request) { handlerNodeA10v2Health(w, r, "/v1/at2/healthcheck/") })
	register("/v1/at3/node/", func(w http.ResponseWriter, r *http.Request) { handlerNodeA10v3(w, r, "/v1/at3/node/") })

	if tls {
		log.Printf("serving HTTPS on TCP %s LISTEN=[%s] TLS=%v", addr, os.Getenv("LISTEN"), tls)
		if err := listenAndServeTLS(addr, cert, key, nil, true); err != nil {
			log.Fatalf("listenAndServeTLS: %s: %v", addr, err)
		}
		return
	}

	log.Printf("serving HTTP on TCP %s LISTEN=[%s] TLS=%v", addr, os.Getenv("LISTEN"), tls)
	if err := listenAndServe(addr, nil, true); err != nil {
		log.Fatalf("listenAndServe: %s: %v", addr, err)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func listenAndServeTLS(addr, certFile, keyFile string, handler http.Handler, keepalive bool) error {
	server := &http.Server{Addr: addr, Handler: handler}
	server.SetKeepAlivesEnabled(keepalive)
	return server.ListenAndServeTLS(certFile, keyFile)
}

func listenAndServe(addr string, handler http.Handler, keepalive bool) error {
	server := &http.Server{Addr: addr, Handler: handler}
	server.SetKeepAlivesEnabled(keepalive)
	return server.ListenAndServe()
}

type handlerFunc func(w http.ResponseWriter, r *http.Request)

func register(path string, handler handlerFunc) {
	log.Printf("registering path: [%s]", path)
	http.HandleFunc(path, handler)
}

func handlerRoot(w http.ResponseWriter, r *http.Request, path string) {

	if r.URL.Path != path {
		sendNotFound("handlerRoot", w, r)
		return
	}

	msg := fmt.Sprintf("handlerRoot: method=%s url=%s from=%s", r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	nothing := fmt.Sprintf("nothing to see here: [%s]", r.URL.Path)

	io.WriteString(w, nothing+"\n")
}

func hidePassword(pwd string) string {
	if showPasswords {
		return pwd
	}
	return "<pwd-hidden>"
}
