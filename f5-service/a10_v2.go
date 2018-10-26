package main

import (
	"encoding/json"
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
// Source: https://github.com/a10networks/tps-scripts/blob/master/axapi_curl_example.txt
//
// curl -k -X POST -H 'content-type: application/json' -d '{"credentials": {"username": "admin", "password": "a10"}}' 'https://192.168.199.152/axapi/v3/auth'
//
// V2:
//
// Source: https://www.a10networks.com/resources/articles/axapi-python
//
// https://10.255.255.6/services/rest/V2/?method=authenticate&username=admin&password=a10&format=json
//
// V2.1:
//
// Source: https://github.com/a10networks/acos-client/blob/master/acos_client/v21/session.py
//
// url:       /services/rest/v2.1/?format=json&method=authenticate
// post body: { "username": username, "password": password }

func a10v21url(host, method string) string {
	return "https://" + host + "/services/rest/v2.1/?format=json&method=" + method
}

func a10v21urlSession(host, method, sessionId string) string {
	return a10v21url(host, method) + "&session_id=" + sessionId
}

func nodeA10v2RuleGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2RuleGet"

	host := fields[0]

	sessionId := nodeA10v2Auth(w, r, host, username, password)
	if sessionId == "" {
		return
	}

	vsList := a10VirtualServerList(host, sessionId)
	var list1 string
	for _, vs := range vsList {
		msg := fmt.Sprintf("virtual server: %v", vs)
		list1 += msg + "\n"
		log.Print(msg)
	}

	sgList := a10ServiceGroupList(host, sessionId)
	var list2 string
	for _, vs := range sgList {
		msg := fmt.Sprintf("service group: %v", vs)
		list2 += msg + "\n"
		log.Print(msg)
	}

	sList := a10ServerList(host, sessionId)
	var list3 string
	for _, s := range sList {
		msg := fmt.Sprintf("server: %v", s)
		list3 += msg + "\n"
		log.Print(msg)
	}

	if errClose := nodeA10v2Close(w, r, host, sessionId); errClose != nil {
		log.Printf(me+": method=%s url=%s from=%s close session_id=[%s] error: %v", r.Method, r.URL.Path, r.RemoteAddr, sessionId, errClose)
		// log warning only
	}

	writeStr(me, w, "done: "+list1+list2+list3)
}

type a10VServer struct {
	name         string
	address      string
	port         string
	serviceGroup string
}

type a10ServiceGroup struct {
	name    string
	members []string
}

type a10Server struct {
	name  string
	host  string
	ports []string
}

func mapGetStr(tab map[string]interface{}, key string) string {
	value, found := tab[key]
	if !found {
		log.Printf("mapGetStr: key=[%s] not found", key)
		return ""
	}
	str, isStr := value.(string)
	if !isStr {
		log.Printf("mapGetStr: key=[%s] non-string value: [%v]", key, value)
		return ""
	}
	return str
}

func a10ServerList(host, sessionId string) []a10Server {
	var list []a10Server

	servers, errGet := a10SessionGet(host, "slb.server.getAll", sessionId)
	if errGet != nil {
		return list
	}

	//log.Printf("servers: [%s]", string(servers))

	sList := jsonExtractList(servers, "server_list")
	if sList == nil {
		return list
	}

	for _, s := range sList {
		sMap, isMap := s.(map[string]interface{})
		if !isMap {
			continue
		}

		name := mapGetStr(sMap, "name")
		host := mapGetStr(sMap, "host")
		//host := "address=fixme"
		server := a10Server{name: name, host: host}

		portList := sMap["port_list"]
		pList, isList := portList.([]interface{})
		if !isList {
			continue
		}
		for _, p := range pList {
			pMap, isPMap := p.(map[string]interface{})
			if !isPMap {
				continue
			}
			portNum := pMap["port_num"]
			server.ports = append(server.ports, fmt.Sprintf("%v", portNum))
		}

		list = append(list, server)
	}

	return list
}

func a10ServiceGroupList(host, sessionId string) []a10ServiceGroup {
	var list []a10ServiceGroup

	groups, errGet := a10SessionGet(host, "slb.service_group.getAll", sessionId)
	if errGet != nil {
		return list
	}

	log.Printf("groups: [%s]", string(groups))

	sgList := jsonExtractList(groups, "service_group_list")
	if sgList == nil {
		return list
	}

	for _, sg := range sgList {
		sgMap, isMap := sg.(map[string]interface{})
		if !isMap {
			continue
		}

		name := mapGetStr(sgMap, "name")
		group := a10ServiceGroup{name: name}

		memberList := sgMap["member_list"]
		mList, isList := memberList.([]interface{})
		if !isList {
			continue
		}
		for _, m := range mList {
			group.members = append(group.members, fmt.Sprintf("%v", m))
		}

		list = append(list, group)
	}

	return list
}

func a10VirtualServerList(host, sessionId string) []a10VServer {
	var list []a10VServer

	bodyVirtServers, errGet := a10SessionGet(host, "slb.virtual_server.getAll", sessionId)
	if errGet != nil {
		return list
	}

	vsList := jsonExtractList(bodyVirtServers, "virtual_server_list")
	if vsList == nil {
		return list
	}

	for _, vs := range vsList {
		vsMap, isMap := vs.(map[string]interface{})
		if !isMap {
			continue
		}

		name := mapGetStr(vsMap, "name")
		addr := mapGetStr(vsMap, "address")
		portList := vsMap["vport_list"]
		pList, isList := portList.([]interface{})
		if !isList {
			continue
		}
		for _, vp := range pList {
			pMap, isPMap := vp.(map[string]interface{})
			if !isPMap {
				continue
			}
			port := pMap["port"]
			pStr := fmt.Sprintf("%v", port)
			sGroup := mapGetStr(pMap, "service_group")
			//log.Printf("virtual server name=[%s] address=[%s] port=[%s] service_group=[%s]", name, addr, pStr, sGroup)
			list = append(list, a10VServer{name, addr, pStr, sGroup})
		}
	}

	return list
}

func jsonExtractList(body []byte, listName string) []interface{} {
	me := "extractList"
	tab := map[string]interface{}{}
	errJson := json.Unmarshal(body, &tab)
	if errJson != nil {
		log.Printf(me+": list=%s json error: %v", listName, errJson)
		return nil
	}
	list, found := tab[listName]
	if !found {
		log.Printf(me+": list=%s not found", listName)
		return nil
	}
	slice, isSlice := list.([]interface{})
	if !isSlice {
		log.Printf(me+": list=%s not an slice", listName)
		return nil
	}
	return slice
}

func a10SessionGet(host, method, sessionId string) ([]byte, error) {
	me := "a10SessionGet"
	api := a10v21urlSession(host, method, sessionId)
	body, err := httpGet(api)
	if err != nil {
		log.Printf(me+": api=[%s] error: %v", api, err)
	}
	return body, err
}

func nodeA10v2Close(w http.ResponseWriter, r *http.Request, host, sessionId string) error {

	me := "nodeA10v2Close"

	api := a10v21urlSession(host, "session.close", sessionId)

	format := `{"session_id": "%s"}`
	payload := fmt.Sprintf(format, sessionId)

	log.Printf(me+": method=%s url=%s from=%s session_id=[%s] api=%s payload=[%s] closing", r.Method, r.URL.Path, r.RemoteAddr, sessionId, api, payload)

	_, errPost := httpPostString(api, "application/json", payload)

	if errPost == nil {
		log.Printf(me+": method=%s url=%s from=%s session_id=[%s] api=%s payload=[%s] closed", r.Method, r.URL.Path, r.RemoteAddr, sessionId, api, payload)
	}

	return errPost
}

func nodeA10v2Auth(w http.ResponseWriter, r *http.Request, host, username, password string) string {

	me := "nodeA10v2Auth"

	//body, errAuth := a10v2auth(r, host, username, password)
	body, errAuth := a10v21auth(r, host, username, password)
	if errAuth != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errAuth)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
		return ""
	}

	response := map[string]interface{}{}

	errJson := json.Unmarshal(body, &response)
	if errJson != nil {
		log.Printf(me+": method=%s url=%s from=%s auth json: %v", r.Method, r.URL.Path, r.RemoteAddr, errJson)
		http.Error(w, host+" bad gateway - auth json", http.StatusBadGateway) // 502
		return ""
	}

	id, found := response["session_id"]
	if !found {
		log.Printf(me+": method=%s url=%s from=%s missing session_id", r.Method, r.URL.Path, r.RemoteAddr)
		http.Error(w, host+" bad gateway - auth missing session_id", http.StatusBadGateway) // 502
		return ""
	}

	session_id, isStr := id.(string)
	if !isStr {
		log.Printf(me+": method=%s url=%s from=%s missing session_id", r.Method, r.URL.Path, r.RemoteAddr)
		http.Error(w, host+" bad gateway - auth session_id not a string", http.StatusBadGateway) // 502
		return ""
	}

	log.Printf(me+": method=%s url=%s from=%s session_id=[%s]", r.Method, r.URL.Path, r.RemoteAddr, session_id)

	return session_id
}

func a10v2auth(r *http.Request, host, username, password string) ([]byte, error) {

	me := "a10v2auth"

	a10host := "https://" + host

	// Attention: this is a V2 api, do not use V21 helpers
	format := "/services/rest/V2/?method=authenticate&username=%s&password=%s&format=json"
	api := a10host + fmt.Sprintf(format, username, password)                  // real path
	apiLog := a10host + fmt.Sprintf(format, username, hidePassword(password)) // path used for logging (hide password)

	log.Printf(me+": method=%s url=%s from=%s opening: %s", r.Method, r.URL.Path, r.RemoteAddr, apiLog)

	return httpGet(api)
}

func a10v21auth(r *http.Request, host, username, password string) ([]byte, error) {

	me := "a10v21auth"

	api := a10v21url(host, "authenticate")

	format := `{ "username": "%s", "password": "%s" }`
	payload := fmt.Sprintf(format, username, password)                  // real payload
	payloadLog := fmt.Sprintf(format, username, hidePassword(password)) // payload used for logging (hide password)

	log.Printf(me+": method=%s url=%s from=%s opening=%s payload=[%s]", r.Method, r.URL.Path, r.RemoteAddr, api, payloadLog)

	return httpPostString(api, "application/json", payload)
}
