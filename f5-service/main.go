package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
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

	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	username, password, authOK := r.BasicAuth()
	if !authOK {
		http.Error(w, "not authorized", http.StatusUnauthorized) // 401
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, path)

	msg := fmt.Sprintf("handlerNodeF5: method=%s url=%s from=%s suffix=[%s] authUser=[%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix, username)
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
		nodeF5RuleGet(w, r, username, password, fields)
	case http.MethodPost:
		nodeF5RulePost(w, r, username, password, fields)
	case http.MethodDelete:
		nodeF5RuleDelete(w, r, username, password, fields)
	default:
		w.Header().Set("Allow", "POST") // required by 405 error
		http.Error(w, r.Method+" method not supported", 405)
		sendNotSupported("handlerNodeF5", w, r)
	}
}

func nodeF5RuleGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	host := fields[0]

	f5Host := "https://" + host

	log.Printf("nodeF5RuleGet: method=%s url=%s from=%s f5.NewBasicClient opening: %s", r.Method, r.URL.Path, r.RemoteAddr, f5Host)

	f5Client, errOpen := f5.NewBasicClient(f5Host, username, password)

	if errOpen != nil {
		log.Printf("nodeF5RuleGet: method=%s url=%s from=%s f5.NewBasicClient: %v", r.Method, r.URL.Path, r.RemoteAddr, errOpen)
		http.Error(w, host+" bad gateway - open", http.StatusBadGateway) // 502
		return
	}

	f5Client.DisableCertCheck()

	ltmClient := ltm.New(f5Client)

	vsConfigList, errVirtList := ltmClient.Virtual().ListAll()
	if errVirtList != nil {
		log.Printf("nodeF5RuleGet: method=%s url=%s from=%s virtual list: %v", r.Method, r.URL.Path, r.RemoteAddr, errVirtList)
		http.Error(w, host+" bad gateway - virtual list", http.StatusBadGateway) // 502
		return
	}

	/*
		poolMembers := ltmClient.PoolMembers()
		members, errMembersList := poolMembers.ListAll()
		if errMembersList != nil {
			log.Printf("nodeF5RuleGet: method=%s url=%s from=%s pool members list: %v", r.Method, r.URL.Path, r.RemoteAddr, errMembersList)
			http.Error(w, host+" bad gateway - pool members list", http.StatusBadGateway) // 502
			return
		}
	*/

	poolClient := ltmClient.Pool()
	poolList, errPoolList := poolClient.ListAll()
	if errPoolList != nil {
		log.Printf("nodeF5RuleGet: method=%s url=%s from=%s pool list: %v", r.Method, r.URL.Path, r.RemoteAddr, errPoolList)
		http.Error(w, host+" bad gateway - pool list", http.StatusBadGateway) // 502
		return
	}

	node := ltmClient.Node()
	nodes, errNodesList := node.ListAll()
	if errNodesList != nil {
		log.Printf("nodeF5RuleGet: method=%s url=%s from=%s nodes list: %v", r.Method, r.URL.Path, r.RemoteAddr, errNodesList)
		http.Error(w, host+" bad gateway - nodes list", http.StatusBadGateway) // 502
		return
	}

	for _, v := range vsConfigList.Items {
		log.Printf("virtual server: user=%s virtual=%s destination=%s pool=%s partition=%s", username, v.Name, v.Destination, v.Pool, v.Partition)
	}

	/*
		for _, m := range members.Items {
			log.Printf("pool member: user=%s pool=%s partition=%s", username, m.Name, m.Partition)
		}
	*/

	for _, p := range poolList.Items {
		log.Printf("pool: user=%s name=%s members=%s", username, p.Name, p.Members)
	}

	/*
		for _, n := range nodes.Items {
			log.Printf("node: user=%s name=%s address=%s partition=%s session=%s state=%s", username, n.Name, n.Address, n.Partition, n.Session, n.State)
		}
	*/

	//writeStr("nodeF5RuleGet", w, fmt.Sprintf("vs:\n%v\nmembers:\n%v\nnodes:\n%v\n", vsConfigList, members, nodes))
	writeStr("nodeF5RuleGet", w, fmt.Sprintf("vs:\n%v\nnodes:\n%v\n", vsConfigList, nodes))
}

func writeStr(caller string, w http.ResponseWriter, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		log.Printf("%s writeStr: %v", caller, err)
	}
}

func nodeF5RulePost(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	host := fields[0]
	sendNotImplemented("nodeF5RulePost:FIXME-WRITEME:"+host, w, r)
}

func nodeF5RuleDelete(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	host := fields[0]
	if len(fields) < 3 {
		reason := fmt.Sprintf("missing path fields: %d < %d", len(fields), 3)
		sendBadRequest("nodeF5RuleDelete", reason, w, r)
		return
	}
	ruleName := fields[2]
	sendNotImplemented("nodeF5RuleDelete:FIXME-WRITEME:host="+host+":rule="+ruleName, w, r)
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
