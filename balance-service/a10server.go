package main

import (
	"net/http"
	//"github.com/udhos/a10-go-rest-client/a10go"
)

// /v1/at/node/<host>/server/
// ^^^^^^^^^^^^
// prefix

func nodeA10v2Server(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2Server"

	switch r.Method {
	case http.MethodGet:
		nodeA10v2ServerGet(w, r, username, password, fields)
	case http.MethodPut:
		nodeA10v2ServerPut(debug, dry, w, r, username, password, fields)
	default:
		sendNotSupported(me, w, r)
	}
}

func nodeA10v2ServerGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2ServerGet"
	writeStr(me, w, "server GET hello\n")
}

func nodeA10v2ServerPut(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2ServerPut"
	writeStr(me, w, "server PUT hello\n")
}
