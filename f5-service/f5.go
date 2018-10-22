package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
)

// /v1/ff/node/<host>/rule/<rule>
// ^^^^^^^^^^^^
// prefix
func handlerNodeF5(w http.ResponseWriter, r *http.Request, path string) {

	me := "handlerNodeF5"

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
	log.Printf(me+": method=%s url=%s from=%s suffix=[%s] auth realm=[%s] auth=[%s:%s]", r.Method, r.URL.Path, r.RemoteAddr, suffix, realm, username, password)

	ruleField := fields[1]
	if ruleField != "rule" {
		reason := fmt.Sprintf("missing fule field: [%s]", ruleField)
		sendBadRequest(me, reason, w, r)
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
		sendNotSupported(me, w, r)
	}
}

func nodeF5RuleGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeF5RuleGet"

	host := fields[0]

	f5Host := "https://" + host

	log.Printf(me+": method=%s url=%s from=%s f5.NewBasicClient opening: %s", r.Method, r.URL.Path, r.RemoteAddr, f5Host)

	f5Client, errOpen := f5.NewBasicClient(f5Host, username, password)

	if errOpen != nil {
		log.Printf(me+": method=%s url=%s from=%s f5.NewBasicClient: %v", r.Method, r.URL.Path, r.RemoteAddr, errOpen)
		http.Error(w, host+" bad gateway - open", http.StatusBadGateway) // 502
		return
	}

	f5Client.DisableCertCheck()

	ltmClient := ltm.New(f5Client)

	vsConfigList, errVirtList := ltmClient.Virtual().ListAll()
	if errVirtList != nil {
		log.Printf(me+": method=%s url=%s from=%s virtual list: %v", r.Method, r.URL.Path, r.RemoteAddr, errVirtList)
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
		log.Printf(me+": method=%s url=%s from=%s pool list: %v", r.Method, r.URL.Path, r.RemoteAddr, errPoolList)
		http.Error(w, host+" bad gateway - pool list", http.StatusBadGateway) // 502
		return
	}

	node := ltmClient.Node()
	nodes, errNodesList := node.ListAll()
	if errNodesList != nil {
		log.Printf(me+": method=%s url=%s from=%s nodes list: %v", r.Method, r.URL.Path, r.RemoteAddr, errNodesList)
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

	//writeStr(me, w, fmt.Sprintf("vs:\n%v\nmembers:\n%v\nnodes:\n%v\n", vsConfigList, members, nodes))
	writeStr(me, w, fmt.Sprintf("vs:\n%v\nnodes:\n%v\n", vsConfigList, nodes))
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
