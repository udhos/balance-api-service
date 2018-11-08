package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/sanity-io/litter"
	"github.com/udhos/a10-go-rest-client/a10go"
)

type backend struct {
	VirtualServerName    string
	VirtualServerAddress string
	VirtualServerPorts   []backendPort
	ServiceGroups        []backendServiceGroup
	BackendName          string
	BackendAddress       string
	BackendPorts         []backendPort
}

type backendServiceGroup struct {
	Name     string
	Protocol string
}

type backendPort struct {
	Port     string
	Protocol string
}

// /v1/at/node/<host>/backend/
// ^^^^^^^^^^^^
// prefix

func nodeA10v2Backend(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {

	me := "nodeA10v2Backend"

	switch r.Method {
	case http.MethodGet:
		nodeA10v2BackendGet(w, r, username, password, fields)
	case http.MethodDelete:
		nodeA10v2BackendDelete(debug, dry, w, r, username, password, fields)
	case http.MethodPut:
		nodeA10v2BackendPut(debug, dry, w, r, username, password, fields)
	default:
		sendNotSupported(me, w, r)
	}
}

func nodeA10v2BackendGet(w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2BackendGet"

	host := fields[0]

	c := a10go.New(host, a10go.Options{})

	errLogin := c.Login(username, password)
	if errLogin != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errLogin)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
	}

	backendTab := fetchBackendTable(c)

	if errClose := c.Logout(); errClose != nil {
		log.Printf(me+": method=%s url=%s from=%s close error: %v", r.Method, r.URL.Path, r.RemoteAddr, errClose)
		// log warning only
	}

	sendBackendList(me, w, r, backendTab)
}

func sendBackendList(me string, w http.ResponseWriter, r *http.Request, tab map[string]*backend) {

	list := []*backend{}

	for _, b := range tab {
		list = append(list, b)
	}

	query := r.URL.Query()
	if _, found := query["debug"]; found {
		writeStr(me, w, litter.Sdump(list))
		writeLine(me, w)
		return
	}

	buf, errMarshal := json.MarshalIndent(list, "", " ")
	if errMarshal != nil {
		log.Printf(me+": method=%s url=%s from=%s json error: %v", r.Method, r.URL.Path, r.RemoteAddr, errMarshal)
		sendInternalError(me, w, r) // http 500
		return
	}
	writeBuf(me, w, buf)
	writeLine(me, w)
}

func nodeA10v2BackendDelete(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2BackendDelete"
	log.Printf(me + " FIXME WRITEME")
	writeStr(me, w, "backend DELETE hello\n")
}

func nodeA10v2BackendPut(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2BackendPut"
	log.Printf(me + " FIXME WRITEME")
	writeStr(me, w, "backend PUT hello\n")
}

func fetchBackendTable(c *a10go.Client) map[string]*backend {
	backendTab := map[string]*backend{}

	// collect all information from A10
	sList := c.ServerList()
	vsList := c.VirtualServerList()
	sgList := c.ServiceGroupList()

	// build backend table
	for _, s := range sList {
		b := backend{
			BackendName:    s.Name,
			BackendAddress: s.Host,
		}
		for _, p := range s.Ports {
			b.BackendPorts = append(b.BackendPorts, backendPort{Port: p.Number, Protocol: A10ProtocolName(p.Protocol)})
		}
		backendTab[b.BackendName] = &b
	}

	backendUniqueGroups := map[string]struct{}{} // prevent repeatedly adding group as backend parent

	// scan service group table
	// this loop IS able to find all service groups (including those ones detached from virtual servers)
	groupTab := map[string]a10go.A10ServiceGroup{}
	for _, sg := range sgList {
		groupTab[sg.Name] = sg // record group info by name for below
		for _, sgm := range sg.Members {
			b, found := backendTab[sgm.Name]
			log.Printf("fetchBackendTable: group=%s backend=%s found=%v", sg.Name, sgm.Name, found)
			if !found {
				continue
			}
			dedupKey := b.BackendName + " " + sg.Name
			_, dup := backendUniqueGroups[dedupKey]
			log.Printf("fetchBackendTable: key=[%s] group=%s backend=%s dup=%v", dedupKey, sg.Name, sgm.Name, dup)
			if dup {
				continue
			}
			backendUniqueGroups[dedupKey] = struct{}{} // mark as added
			b.ServiceGroups = append(b.ServiceGroups, backendServiceGroup{Name: sg.Name, Protocol: sg.Protocol})
		}
	}

	// scan virtual server list attaching information to backend table
	// this loop is UNABLE to find service groups detached from virtual servers
	for _, vs := range vsList {
		log.Printf("fetchBackendTable: vserver=%s", vs.Name)
		portDedupTab := map[string]struct{}{}
		for _, vp := range vs.VirtualPorts {
			sg, found := groupTab[vp.ServiceGroup]
			if !found {
				log.Printf("fetchBackendTable: vserver=%s group=%s not found", vs.Name, vp.ServiceGroup)
				continue
			}
			portProtoName := A10ProtocolName(vp.Protocol)
			portDedupKey := vp.Port + " " + portProtoName
			for _, sgm := range sg.Members {
				if b, found := backendTab[sgm.Name]; found {
					b.VirtualServerName = vs.Name
					b.VirtualServerAddress = vs.Address
					if _, dup := portDedupTab[portDedupKey]; dup {
						continue
					}
					portDedupTab[portDedupKey] = struct{}{} // mark port as added
					b.VirtualServerPorts = append(b.VirtualServerPorts, backendPort{Port: vp.Port, Protocol: portProtoName})
				}
			}
		}
	}

	return backendTab
}
