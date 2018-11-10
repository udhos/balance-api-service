package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	//"github.com/sanity-io/litter"
	//"github.com/udhos/a10-go-rest-client/a10go"
)

// /v1/at/node/<host>/rule/
// /v1/at/node/<host>/backend/
// ^^^^^^^^^^^^
// prefix
func handlerNodeA10v2(debug, dry bool, w http.ResponseWriter, r *http.Request, path string) {

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

	log.Printf("handlerNodeA10v2: FIXME? Access-Control-Allow-Origin")
	w.Header().Set("Access-Control-Allow-Origin", "*") // FIXME?

	optionField := fields[1]
	switch optionField {
	/*
		case "rule":
			nodeA10v2Rule(debug, dry, w, r, username, password, fields)
	*/
	case "backend":
		nodeA10v2Backend(debug, dry, w, r, username, password, fields)
	default:
		reason := fmt.Sprintf("missing option field: [%s]", optionField)
		sendBadRequest(me, reason, w, r)
	}
}

/*
func nodeA10v2Rule(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2Rule"

	switch r.Method {
	case http.MethodGet:
		nodeA10v2RuleGet(w, r, username, password, fields)
	case http.MethodPut:
		nodeA10v2RulePut(debug, dry, w, r, username, password, fields)
	default:
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

func nodeA10v2RulePut(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

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

	log.Printf("nodeA10v2RulePut: debug=%v", debug)
	c := a10go.New(host, a10go.Options{Debug: debug, Dry: dry})

	errLogin := c.Login(username, password)
	if errLogin != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errLogin)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
	}

	oldList := fetchVirtualList(c) // oldList: before change

	log.Printf("oldList: %v", oldList)

	// newList: perform change here

	put(c, oldList, newList)

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

	log.Printf("fetchVirtualList: total: vServers=%d groups=%d servers=%d", len(vsList), len(sgList), len(sList))

	var countGroups, countServers int

	vList := []virtual{}
	for _, vs := range vsList {
		v := virtual{Name: vs.Name, Address: vs.Address}

		//log.Printf("fetchVirtualList: v=%s", vs.Name)

		for _, vp := range vs.VirtualPorts {

			//log.Printf("fetchVirtualList: v=%s p=%s", vs.Name, vp.Port)

			for _, sg := range sgList {
				if sg.Name != vp.ServiceGroup {
					continue
				}

				p := pool{Name: vp.ServiceGroup, Port: vp.Port, Protocol: vp.Protocol}

				//log.Printf("fetchVirtualList: v=%s p=%s g=%s", vs.Name, vp.Port, sg.Name)

				for _, sgm := range sg.Members {
					for _, s := range sList {
						if sgm.Name != s.Name {
							continue
						}

						//log.Printf("fetchVirtualList: v=%s p=%s g=%s m=%s", vs.Name, vp.Port, sg.Name, sgm.Name)

						host := server{Name: s.Name, Address: s.Host}
						for _, port := range s.Ports {
							//log.Printf("fetchVirtualList: v=%s p=%s g=%s m=%s p=%s", vs.Name, vp.Port, sg.Name, sgm.Name, port.Number)
							protoName := A10ProtocolName(port.Protocol)
							host.Ports = append(host.Ports, serverPort{Port: port.Number, Protocol: protoName})
						}
						p.Members = append(p.Members, host)
						countServers++
					}
				}

				v.Pools = append(v.Pools, p)
				countGroups++
			}
		}

		vList = append(vList, v)
	}

	log.Printf("fetchVirtualList: linked: vServers=%d groups=%d servers=%d", len(vsList), countGroups, countServers)

	return vList
}
*/
