package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sanity-io/litter"
	"github.com/udhos/a10-go-rest-client/a10go"
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
	case http.MethodPut:
		nodeA10v2RulePut(w, r, username, password, fields)
	default:
		w.Header().Set("Allow", "POST") // required by 405 error
		http.Error(w, r.Method+" method not supported", 405)
		sendNotSupported(me, w, r)
	}
}

func nodeA10v2RuleGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2RuleGet"

	host := fields[0]

	c := a10go.New(host, a10go.Options{})

	errLogin := c.Login(username, password)
	if errLogin != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errLogin)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
	}

	vList := fetchVirtualList(c)

	if errClose := c.Logout(); errClose != nil {
		log.Printf(me+": method=%s url=%s from=%s close error: %v", r.Method, r.URL.Path, r.RemoteAddr, errClose)
		// log warning only
	}

	sendVirtualList(me, w, r, vList)
}

func sendVirtualList(me string, w http.ResponseWriter, r *http.Request, vList []virtual) {

	query := r.URL.Query()
	if _, found := query["debug"]; found {
		writeStr(me, w, litter.Sdump(vList))
		return
	}

	buf, errMarshal := json.MarshalIndent(vList, "", " ")
	if errMarshal != nil {
		log.Printf(me+": method=%s url=%s from=%s json error: %v", r.Method, r.URL.Path, r.RemoteAddr, errMarshal)
		sendInternalError(me, w, r) // http 500
		return
	}
	writeBuf(me, w, buf)
}

func nodeA10v2RulePut(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2RulePut"

	var newList []virtual

	dec := json.NewDecoder(r.Body)

	errJson := dec.Decode(&newList)
	if errJson != nil {
		reason := fmt.Sprintf("json error: %v", errJson)
		sendBadRequest(me, reason, w, r)
		return
	}

	log.Printf("newList: %v", newList)

	host := fields[0]

	c := a10go.New(host, a10go.Options{})

	errLogin := c.Login(username, password)
	if errLogin != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errLogin)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
	}

	oldList := fetchVirtualList(c) // oldList: before change

	log.Printf("oldList: %v", oldList)

	// newList: perform change here

	finalList := fetchVirtualList(c) // finalList: after change

	if errClose := c.Logout(); errClose != nil {
		log.Printf(me+": method=%s url=%s from=%s close error: %v", r.Method, r.URL.Path, r.RemoteAddr, errClose)
		// log warning only
	}

	sendVirtualList(me, w, r, finalList)
}

func fetchVirtualList(c *a10go.Client) []virtual {

	vsList := c.VirtualServerList()
	sgList := c.ServiceGroupList()
	sList := c.ServerList()

	vList := []virtual{}
	for _, vs := range vsList {
		v := virtual{Name: vs.Name, Address: vs.Address, Port: vs.Port}

		for _, vsg := range vs.ServiceGroups {

			for _, sg := range sgList {
				if sg.Name != vsg {
					continue
				}

				p := pool{Name: vsg}

				for _, sgm := range sg.Members {
					for _, s := range sList {
						if sgm.Name != s.Name {
							continue
						}
						for _, port := range s.Ports {
							protoName := "unknown"
							switch port.Protocol {
							case "2":
								protoName = "tcp"
							case "3":
								protoName = "tcp"
							default:
								protoName = "unknown:" + port.Protocol
							}
							p.Members = append(p.Members, server{Name: s.Name, Address: s.Host, Port: port.Number, Protocol: protoName})
						}
					}
				}

				v.Pools = append(v.Pools, p)
			}
		}

		vList = append(vList, v)
	}

	return vList
}
