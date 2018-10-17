package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
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

	register("/v1/ff/node/", func(w http.ResponseWriter, r *http.Request) { handlerNodeF5(w, r, "/v1/ff/node/") })
	register("/v1/at/node/", func(w http.ResponseWriter, r *http.Request) { handlerNodeA10(w, r, "/v1/at/node/") })

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

func handlerNodeA10(w http.ResponseWriter, r *http.Request, path string) {
	msg := fmt.Sprintf("handlerNodeA10: method=%s url=%s from=%s - NOT IMPLEMENTED", r.Method, r.URL.Path, r.RemoteAddr)
	log.Print(msg)

	notImplemented := fmt.Sprintf("not implemented: [%s]", r.URL.Path)

	w.WriteHeader(http.StatusNotImplemented)

	io.WriteString(w, notImplemented+"\n")
}

// /v1/ff/node/<host>/rule/<rule>
// ^^^^^^^^^^^^
// prefix
func handlerNodeF5(w http.ResponseWriter, r *http.Request, path string) {

	if !strings.HasPrefix(r.URL.Path, path) {
		sendNotFound("handlerNodeF5", w, r)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, path)

	msg := fmt.Sprintf("handlerNodeF5: method=%s url=%s from=%s suffix=[%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix)
	log.Print(msg)

	fields := strings.FieldsFunc(suffix, func(r rune) bool { return r == '/' })

	if len(fields) < 2 {
		reason := fmt.Sprintf("missing path fields: %d < %d", len(fields), 2)
		sendBadRequest("handlerNodeF5", reason, w, r)
		return
	}

	ruleField := fields[1]
	if ruleField != "rule" {
		reason := fmt.Sprintf("missing fule field: [%s]", ruleField)
		sendBadRequest("handlerNodeF5", reason, w, r)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*") // FIXME??

	switch r.Method {
	case http.MethodGet:
		nodeF5RuleGet(w, r, fields)
	case http.MethodPost:
		nodeF5RulePost(w, r, fields)
	case http.MethodDelete:
		nodeF5RuleDelete(w, r, fields)
	default:
		w.Header().Set("Allow", "POST") // required by 405 error
		http.Error(w, r.Method+" method not supported", 405)
		sendNotSupported("handlerNodeF5", w, r)
	}
}

func nodeF5RuleGet(w http.ResponseWriter, r *http.Request, fields []string) {
	host := fields[0]
	sendNotImplemented("nodeF5RuleGet:FIXME-WRITEME:"+host, w, r)
}

func nodeF5RulePost(w http.ResponseWriter, r *http.Request, fields []string) {
	host := fields[0]
	sendNotImplemented("nodeF5RulePost:FIXME-WRITEME:"+host, w, r)
}

func nodeF5RuleDelete(w http.ResponseWriter, r *http.Request, fields []string) {
	host := fields[0]
	if len(fields) < 3 {
		reason := fmt.Sprintf("missing path fields: %d < %d", len(fields), 3)
		sendBadRequest("nodeF5RuleDelete", reason, w, r)
		return
	}
	sendNotImplemented("nodeF5RuleDelete:FIXME-WRITEME:"+host, w, r)
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
