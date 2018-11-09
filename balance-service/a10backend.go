package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sanity-io/litter"
	"github.com/udhos/a10-go-rest-client/a10go"
)

// backend is the main type for the /backend/ route
// A10 full backend path is:
// virtual server -> virtual port (within virtual server) -> service group -> backend
type backend struct {
	VirtualServers []backendVirtualServer
	ServiceGroups  []backendServiceGroup
	BackendName    string
	BackendAddress string
	BackendPorts   []backendPort
}

// FIXME:
// bvs for for backend1 should be distinct from bvs for backend2
// because backend1 should report in its parent VS only VirtualPorts from its path
type backendVirtualServer struct {
	Name         string
	Address      string
	VirtualPorts []backendPort
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
		return
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

	var be backend

	dec := json.NewDecoder(r.Body)

	errJson := dec.Decode(&be)
	if errJson != nil {
		reason := fmt.Sprintf("json error: %v", errJson)
		sendBadRequest(me, reason, w, r)
		return
	}

	if be.BackendName == "" {
		sendBadRequest(me, "missing backend name", w, r)
		return
	}

	log.Printf("%s: backend=[%s] serviceGroups=%d", me, be.BackendName, len(be.ServiceGroups))

	host := fields[0]
	c := a10go.New(host, a10go.Options{Debug: debug})

	errLogin := c.Login(username, password)
	if errLogin != nil {
		log.Printf(me+": method=%s url=%s from=%s auth: %v", r.Method, r.URL.Path, r.RemoteAddr, errLogin)
		http.Error(w, host+" bad gateway - auth", http.StatusBadGateway) // 502
		return
	}

	defer func() {
		if errClose := c.Logout(); errClose != nil {
			log.Printf(me+": method=%s url=%s from=%s close error: %v", r.Method, r.URL.Path, r.RemoteAddr, errClose)
			// log warning only
		}
	}()

	if len(be.ServiceGroups) < 1 {
		// service groups not provided - delete unlinked server
		errDelete := c.ServerDelete(be.BackendName)
		if errDelete != nil {
			log.Printf(me+": method=%s url=%s from=%s delete server: %v", r.Method, r.URL.Path, r.RemoteAddr, errDelete)
			http.Error(w, host+" bad gateway - delete server", http.StatusBadGateway) // 502
			return
		}
		writeStr(me, w, "server deleted\n")
	} else {
		// service groups provided - unlink server from groups
		writeStr(me, w, "server unlinked\n")
	}
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
			if !found {
				continue
			}
			dedupKey := b.BackendName + " " + sg.Name
			_, dup := backendUniqueGroups[dedupKey]
			if dup {
				continue
			}
			backendUniqueGroups[dedupKey] = struct{}{} // mark as added
			b.ServiceGroups = append(b.ServiceGroups, backendServiceGroup{Name: sg.Name, Protocol: sg.Protocol})
		}
	}

	// bvsTab: record virtual server ports
	bvsTab := map[string]*backendVirtualServer{} // bvsName => bvs
	for _, vs := range vsList {
		bvs := &backendVirtualServer{Name: vs.Name, Address: vs.Address}
		for _, vp := range vs.VirtualPorts {
			bvs.VirtualPorts = append(bvs.VirtualPorts, backendPort{Port: vp.Port, Protocol: A10ProtocolName(vp.Protocol)})
		}
		bvsTab[vs.Name] = bvs
	}

	// scan virtual server list attaching information to backend table
	// this loop is UNABLE to find service groups detached from virtual servers
	for _, vs := range vsList {
		vsDedupTab := map[string]struct{}{} // key=backend
		for _, vp := range vs.VirtualPorts {
			sg, found := groupTab[vp.ServiceGroup]
			if !found {
				log.Printf("fetchBackendTable: vserver=%s group=%s not found", vs.Name, vp.ServiceGroup)
				continue
			}
			for _, sgm := range sg.Members {
				if b, found := backendTab[sgm.Name]; found {

					_, backendFound := vsDedupTab[sgm.Name]
					if backendFound {
						continue // vs already parent of this backend
					}
					bvs, bvsFound := bvsTab[vs.Name]
					if !bvsFound {
						log.Printf("fetchBackendTable: backend=%s vserver=%s NOT FOUND", sgm.Name, vs.Name)
						continue // ugh not possible
					}
					b.VirtualServers = append(b.VirtualServers, *bvs)
					vsDedupTab[sgm.Name] = struct{}{} // mark vs as added as backend parent
				}
			}
		}
	}

	return backendTab
}
