package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sanity-io/litter"
	"github.com/udhos/a10-go-rest-client/a10go"
	"gopkg.in/yaml.v2"
)

// backend is the main type for the /backend/ route
// A10 full backend path is:
// virtual server -> virtual port list -> service group -> list of {backend_name, backend_port} -> server (has own list of ports)

type backend struct {
	VirtualServers []backendVirtualServer
	ServiceGroups  []backendServiceGroup
	BackendName    string
	BackendAddress string
	BackendPorts   []backendPort
}

type backendVirtualServer struct {
	Name         string
	Address      string
	VirtualPorts []backendVirtualPort
}

type backendVirtualPort struct {
	Port         string
	Protocol     string
	ServiceGroup string
}

type backendServiceGroup struct {
	Name     string
	Protocol string
	Members  []backendSGMember // list of members
}

type backendSGMember struct {
	Name string
	Port string
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
		nodeA10v2BackendGet(debug, w, r, username, password, fields)
	case http.MethodDelete:
		nodeA10v2BackendDelete(debug, dry, w, r, username, password, fields)
	case http.MethodPost:
		nodeA10v2BackendPost(debug, dry, w, r, username, password, fields)
	default:
		sendNotSupported(me, w, r)
	}
}

func nodeA10v2BackendGet(debug bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
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

	acceptYAML, _ := clientOptions(debug, r)

	sendBackendList(me, w, r, backendTab, acceptYAML)
}

func sendBackendList(me string, w http.ResponseWriter, r *http.Request, tab map[string]*backend, acceptYAML bool) {

	list := []*backend{}

	for _, b := range tab {
		list = append(list, b)
	}

	// force litter
	query := r.URL.Query()
	if _, found := query["debug"]; found {
		writeStr(me, w, litter.Sdump(list))
		writeLine(me, w)
		return
	}

	// force YAML if supported
	if acceptYAML {
		buf, errMarshal := yaml.Marshal(list)
		if errMarshal != nil {
			log.Printf(me+": method=%s url=%s from=%s yaml error: %v", r.Method, r.URL.Path, r.RemoteAddr, errMarshal)
			sendInternalError(me, w, r) // http 500
			return
		}
		writeBuf(me, w, buf)
		writeLine(me, w)
		return
	}

	// default to JSON
	buf, errMarshal := json.MarshalIndent(list, "", " ")
	if errMarshal != nil {
		log.Printf(me+": method=%s url=%s from=%s json error: %v", r.Method, r.URL.Path, r.RemoteAddr, errMarshal)
		sendInternalError(me, w, r) // http 500
		return
	}
	writeBuf(me, w, buf)
	writeLine(me, w)
}

func decodeBackend(debug bool, body io.Reader, bodyYAML bool, be *backend) error {
	me := "decodeBackend"

	// force YAML if supported
	if bodyYAML {
		log.Printf(me + ": decoding YAML request body")

		buf, errRead := ioutil.ReadAll(body)
		if errRead != nil {
			return fmt.Errorf("read error: %v", errRead)
		}

		errYaml := yaml.Unmarshal(buf, be)
		if errYaml != nil {
			log.Printf(me+": decoding YAML request body - error: %v buf=[%s]", errYaml, string(buf))
			return fmt.Errorf("yaml error: %v", errYaml)
		}

		return nil
	}

	// defaults to JSON
	log.Printf(me + ": decoding JSON request body")
	dec := json.NewDecoder(body)
	errJson := dec.Decode(be)
	if errJson != nil {
		return fmt.Errorf("json error: %v", errJson)
	}

	return nil
}

func decodeRequestBody(debug bool, w http.ResponseWriter, r *http.Request, be *backend) error {

	me := "decodeRequestBody"

	_, bodyYAML := clientOptions(debug, r)

	errDecode := decodeBackend(debug, r.Body, bodyYAML, be)
	if errDecode != nil {
		sendBadRequest(me, errDecode.Error(), w, r)
		return errDecode
	}

	return nil
}

func nodeA10v2BackendDelete(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2BackendDelete"

	var be backend

	if errDecode := decodeRequestBody(debug, w, r, &be); errDecode != nil {
		return
	}

	if be.BackendName == "" {
		sendBadRequest(me, "missing backend name", w, r)
		return
	}

	log.Printf("%s: backend=[%s] serviceGroups=%d", me, be.BackendName, len(be.ServiceGroups))

	host := fields[0]
	c := a10go.New(host, a10go.Options{Debug: debug, Dry: dry})

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

		backendUnlink(c, w, r, be, host)
	}
}

func backendUnlink(c *a10go.Client, w http.ResponseWriter, r *http.Request, be backend, host string) {

	me := "backendUnlink"

	sgList := c.ServiceGroupList() // all available groups

	sgUnlinkList := []a10go.A10ServiceGroup{} // groups linked to backend server

	// find groups linked to backend server
LOOP:
	for _, bsg := range be.ServiceGroups {
		for _, sg := range sgList {
			if sg.Name == bsg.Name {
				// found bsg
				sgUnlinkList = append(sgUnlinkList, sg)
				continue LOOP // next bsg
			}
		}
		// bsg not found
		log.Printf(me+": method=%s url=%s from=%s unlink server: group=%s not found", r.Method, r.URL.Path, r.RemoteAddr, bsg.Name)
		http.Error(w, host+" bad gateway - unlink server: group not found", http.StatusBadGateway) // 502
		return
	}

	log.Printf(me+": backend=[%s] linked groups=%v", be.BackendName, sgUnlinkList)

	var errCount int

	// scan groups unlinking the backend server

	for _, sg := range sgUnlinkList {

		memberList := rebuildMemberList(sg.Name, sg.Members, be)

		// delete previous member list
		errUpdate1 := c.ServiceGroupUpdate(sg.Name, sg.Protocol, nil)
		if errUpdate1 != nil {
			log.Printf(me+": method=%s url=%s from=%s unlink group=%s update-reset: %v", r.Method, r.URL.Path, r.RemoteAddr, sg.Name, errUpdate1)
			errCount++
		}

		// rebuild member list
		errUpdate2 := c.ServiceGroupUpdate(sg.Name, sg.Protocol, memberList)
		if errUpdate2 != nil {
			log.Printf(me+": method=%s url=%s from=%s unlink group=%s update-rebuild: %v", r.Method, r.URL.Path, r.RemoteAddr, sg.Name, errUpdate2)
			errCount++
		}
	}

	writeStr(me, w, fmt.Sprintf("server unlinked - errors:%d\n", errCount))
}

func nodeA10v2BackendPost(debug, dry bool, w http.ResponseWriter, r *http.Request, username, password string, fields []string) {
	me := "nodeA10v2BackendPost"

	var be backend

	if errDecode := decodeRequestBody(debug, w, r, &be); errDecode != nil {
		return
	}

	if be.BackendName == "" {
		sendBadRequest(me, "missing backend name", w, r)
		return
	}

	if be.BackendAddress == "" {
		sendBadRequest(me, "missing backend address", w, r)
		return
	}

	host := fields[0]
	c := a10go.New(host, a10go.Options{Debug: debug, Dry: dry})

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

	// find groups linked to backend server

	sgList := c.ServiceGroupList()        // all available groups
	sgLinked := []a10go.A10ServiceGroup{} // groups to be linked to backend server
LOOP:
	for _, bsg := range be.ServiceGroups {
		for _, sg := range sgList {
			if sg.Name == bsg.Name {
				// found bsg
				sgLinked = append(sgLinked, sg)
				continue LOOP // next bsg
			}
		}
		// bsg not found
		log.Printf(me+": method=%s url=%s from=%s link server: group=%s not found", r.Method, r.URL.Path, r.RemoteAddr, bsg.Name)
		http.Error(w, host+" bad gateway - link server: group not found", http.StatusBadGateway) // 502
		return
	}

	// create or update server?
	sList := c.ServerList()
	var serverFound bool // defaults to create server
	for _, s := range sList {
		if s.Name == be.BackendName {
			serverFound = true // update server
			break
		}
	}

	var portList []string
	for _, p := range be.BackendPorts {
		portList = append(portList, p.Port+","+A10ProtocolNumber(p.Protocol))
	}

	// create or update server

	if serverFound {
		// server exists - update
		errUpdate := c.ServerUpdate(be.BackendName, be.BackendAddress, portList)
		if errUpdate != nil {
			log.Printf(me+": method=%s url=%s from=%s update server: %v", r.Method, r.URL.Path, r.RemoteAddr, errUpdate)
			http.Error(w, host+" bad gateway - update server", http.StatusBadGateway) // 502
			return
		}
	} else {
		// server does not exist - create
		errCreate := c.ServerCreate(be.BackendName, be.BackendAddress, portList)
		if errCreate != nil {
			log.Printf(me+": method=%s url=%s from=%s create server: %v", r.Method, r.URL.Path, r.RemoteAddr, errCreate)
			http.Error(w, host+" bad gateway - create server", http.StatusBadGateway) // 502
			return
		}
	}

	if len(be.ServiceGroups) < 1 {
		// service groups not provided - create/update server
		if serverFound {
			writeStr(me, w, "server updated\n")
		} else {
			writeStr(me, w, "server created\n")
		}
	} else {
		// service groups provided - link server from groups
		backendLink(c, w, r, be, host, sgLinked)
	}
}

// rebuild service group member list excluding
func rebuildMemberList(sgName string, oldMembers []a10go.A10SGMember, be backend) []string {

	me := "rebuildMemberList"

	var memberList []string

	// build service group member list
	for _, m := range oldMembers {
		if m.Name == be.BackendName {
			continue // exclude previous backend server ports from list
		}
		memberList = append(memberList, m.Name+","+m.Port) // keep other existing members
	}

	// append new ports for current backend server
	for _, bsg := range be.ServiceGroups {
		for _, bsgm := range bsg.Members {
			memberList = append(memberList, bsgm.Name+","+bsgm.Port)
		}
	}

	log.Printf(me+": linking group=%s members=%v", sgName, memberList)

	return memberList
}

func backendLink(c *a10go.Client, w http.ResponseWriter, r *http.Request, be backend, host string, sgLinked []a10go.A10ServiceGroup) {

	me := "backendLink"

	var errCount int

	for _, sg := range sgLinked {

		memberList := rebuildMemberList(sg.Name, sg.Members, be)

		errUpdate := c.ServiceGroupUpdate(sg.Name, sg.Protocol, memberList)
		if errUpdate != nil {
			log.Printf(me+": method=%s url=%s from=%s link group=%s: %v", r.Method, r.URL.Path, r.RemoteAddr, sg.Name, errUpdate)
			errCount++
		}
	}

	writeStr(me, w, fmt.Sprintf("server linked - errors:%d\n", errCount))
}

/*
// fetchBackendTableEraseme FIXME ERASEME
func fetchBackendTableEraseme(c *a10go.Client) map[string]*backend {
	backendTab := map[string]*backend{} // backendName => backend

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
	groupTab := map[string]a10go.A10ServiceGroup{} // groupName => group
	for _, sg := range sgList {
		groupTab[sg.Name] = sg // record group info by name for below

		for _, sgm := range sg.Members {
			b, found := backendTab[sgm.Name]
			if !found {
				continue // group backend not found - skip
			}
			dedupKey := b.BackendName + " " + sg.Name
			_, dup := backendUniqueGroups[dedupKey]
			if dup {
				continue
			}
			backendUniqueGroups[dedupKey] = struct{}{} // mark as added

			bsg := backendServiceGroup{Name: sg.Name, Protocol: sg.Protocol}

			b.ServiceGroups = append(b.ServiceGroups, bsg)
		}
	}

	// bvsTab: record virtual server ports
	bvsTab := map[string]*backendVirtualServer{} // bvsName => bvs
	for _, vs := range vsList {
		bvs := &backendVirtualServer{Name: vs.Name, Address: vs.Address}
		for _, vp := range vs.VirtualPorts {
			bvs.VirtualPorts = append(bvs.VirtualPorts, backendVirtualPort{Port: vp.Port, Protocol: A10ProtocolName(vp.Protocol)})
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
*/
