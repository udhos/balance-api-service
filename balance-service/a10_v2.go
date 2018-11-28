package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func isYaml(s string) bool {
	return strings.HasSuffix(s, "yaml") || strings.HasSuffix(s, "*")
}

// Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8

func clientOptions(debug bool, r *http.Request) (acceptYAML, bodyYAML bool) {
	for k, v := range r.Header {
	NEXT_HEADER:
		for _, vv := range v {
			if debug {
				log.Printf("clientOptions: header: [%s]=[%s]", k, vv)
			}
			if k == "Accept" {
				// split on "," -- handle all parts
				typeList := strings.Split(vv, ",")
				for _, t := range typeList {
					// split on ";" -- handle only first part
					typePrefix := strings.Split(t, ";")
					if len(typePrefix) > 0 && isYaml(typePrefix[0]) {
						if debug {
							log.Printf("clientOptions: yaml: Accept: %s", t)
						}
						acceptYAML = true
						continue NEXT_HEADER
					}
				}
			}
			if k == "Content-Type" && isYaml(vv) {
				bodyYAML = true
			}
		}
	}

	if debug {
		log.Printf("clientOptions: acceptYAML=%v bodyYAML=%v", acceptYAML, bodyYAML)
	}

	return
}

// Forwarded: by=<identifier>; for=<identifier>; host=<host>; proto=<http|https>
func forwarded(label string, r *http.Request) (fBy, fFor, fHost, fProto string) {

	me := "forwarded"

LOOP:
	for k, v := range r.Header {
		for _, vv := range v {
			switch k {
			case "X-Forwarded-For":
				fFor = vv
			case "X-Forwarded-Host":
				fHost = vv
			case "X-Forwarded-Proto":
				fProto = vv
			case "Forwarded":
				partList := strings.Split(vv, ";")
				for _, part := range partList {
					p := strings.TrimSpace(part)
					switch {
					case strings.HasPrefix(p, "by="):
						fBy = strings.TrimPrefix(p, "by=")
					case strings.HasPrefix(p, "for="):
						fFor = strings.TrimPrefix(p, "for=")
					case strings.HasPrefix(p, "host="):
						fHost = strings.TrimPrefix(p, "host=")
					case strings.HasPrefix(p, "proto="):
						fProto = strings.TrimPrefix(p, "proto=")
					}
				}
				break LOOP
			}
		}
	}

	log.Printf("%s: %s: by=%s for=%s host=%s proto=%s", label, me, fBy, fFor, fHost, fProto)

	return
}

// /v1/at2/healthcheck
func handlerNodeA10v2Health(w http.ResponseWriter, r *http.Request, path string) {
	writeStr("handlerNodeA10v2Health", w, "health ok\n")
}

// /v1/at2/node/<host>/rule/
// /v1/at2/node/<host>/backend/
// ^^^^^^^^^^^^^
// prefix
func handlerNodeA10v2(debug, dry bool, w http.ResponseWriter, r *http.Request, path string) {

	me := "handlerNodeA10v2"

	if !strings.HasPrefix(r.URL.Path, path) {
		sendNotFound(me, w, r)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, path)

	log.Printf(me+": TLS=%v method=%s url=%s from=%s suffix=[%s]", r.TLS != nil, r.Method, r.URL.Path, r.RemoteAddr, suffix)
	forwarded(me, r)

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
	case "healthcheck":
		writeStr(me, w, "node health ok\n")
	default:
		reason := fmt.Sprintf("unexpected option field: [%s]", optionField)
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
