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

func main() {

	me := os.Args[0]

	log.Printf("%s %s runtime %s GOMAXPROCS=%d", me, version, runtime.Version(), runtime.GOMAXPROCS(0))

	addr := os.Getenv("LISTEN")
	if addr == "" {
		addr = ":8080"
	}

	register("/", func(w http.ResponseWriter, r *http.Request) { handlerRoot(w, r, "/") })
	register("/v1/rule", func(w http.ResponseWriter, r *http.Request) { handlerRule(w, r, "/v1/rule") })
	register("/v1/rule/", func(w http.ResponseWriter, r *http.Request) { handlerRule(w, r, "/v1/rule/") })

	log.Printf("serving HTTP on TCP %s LISTEN=[%s]", addr, os.Getenv("LISTEN"))

	if err := listenAndServe(addr, nil, true); err != nil {
		log.Fatalf("listenAndServe: %s: %v", addr, err)
	}
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

func handlerRule(w http.ResponseWriter, r *http.Request, path string) {

	if r.URL.Path != path {
		sendNotFound("handlerRule", w, r)
		return
	}

	msg := fmt.Sprintf("handlerHello: method=%s url=%s from=%s", r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	hello := "hello rule api"

	io.WriteString(w, hello+"\n")
}

func sendNotFound(label string, w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("%s: method=%s url=%s from=%s - PATH NOT FOUND", label, r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	notFound := fmt.Sprintf("path not found: [%s]", r.URL.Path)

	w.WriteHeader(http.StatusNotFound)

	io.WriteString(w, notFound+"\n")
}
