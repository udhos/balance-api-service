package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// /v1/at/node/<host>/rule/<rule>
// ^^^^^^^^^^^^
// prefix
func handlerNodeA10v2(w http.ResponseWriter, r *http.Request, path string) {

	me := "handlerNodeA10v2"

	if !strings.HasPrefix(r.URL.Path, path) {
		sendNotFound(me, w, r)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, path)

	msg := fmt.Sprintf(me+": method=%s url=%s from=%s suffix=[%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix)
	log.Print(msg)

	fields := strings.FieldsFunc(suffix, func(r rune) bool { return r == '/' })

	if len(fields) < 2 {
		reason := fmt.Sprintf("missing path fields: %d < %d", len(fields), 2)
		sendBadRequest(me, reason, w, r)
		return
	}

	node := fields[0]
	realm := "node-" + node
	log.Printf(me+": method=%s url=%s from=%s suffix=[%s] auth realm=[%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix, realm)
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	username, password, authOK := r.BasicAuth()
	if !authOK {
		http.Error(w, "not authorized", http.StatusUnauthorized) // 401
		return
	}
	log.Printf(me+": method=%s url=%s from=%s suffix=[%s] auth realm=[%s] auth=[%s:%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix, realm, username, hidePassword(password))

	ruleField := fields[1]
	if ruleField != "rule" {
		reason := fmt.Sprintf("missing fule field: [%s]", ruleField)
		sendBadRequest(me, reason, w, r)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*") // FIXME??

	switch r.Method {
	case http.MethodGet:
		nodeA10v2RuleGet(w, r, username, password, fields)
	default:
		w.Header().Set("Allow", "POST") // required by 405 error
		http.Error(w, r.Method+" method not supported", 405)
		sendNotSupported(me, w, r)
	}
}

// V3:
//
// https://github.com/a10networks/tps-scripts/blob/master/axapi_curl_example.txt
//
// curl -k -X POST -H 'content-type: application/json' -d '{"credentials": {"username": "admin", "password": "a10"}}' 'https://192.168.199.152/axapi/v3/auth'
//
// V2:
//
// https://www.a10networks.com/resources/articles/axapi-python
//
// https://10.255.255.6/services/rest/V2/?method=authenticate&username=admin&password=a10&format=json

func nodeA10v2RuleGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2RuleGet"

	host := fields[0]

	a10host := "https://" + host
	format := "/services/rest/V2/?method=authenticate&username=%s&password=%s&format=json"
	api := a10host + fmt.Sprintf(format, username, password)                  // real path
	apiLog := a10host + fmt.Sprintf(format, username, hidePassword(password)) // path used for logging (hide password)

	log.Printf(me+": method=%s url=%s from=%s opening: %s", r.Method, r.URL.Path, r.RemoteAddr, apiLog)

	body, errAuth := httpGet(api)

	if errAuth != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errAuth)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
		return
	}

	writeStr(me, w, "done body:"+string(body))
}
