package main

import (
	//"encoding/json"
	//"fmt"
	//"log"
	//"net/http"

	//"github.com/sanity-io/litter"
	"github.com/udhos/a10-go-rest-client/a10go"
)

func fetchBackendTable2(c *a10go.Client) map[string]*backend {

	// collect all information from A10
	sList := c.ServerList()
	//vsList := c.VirtualServerList()
	sgList := c.ServiceGroupList()

	backendTab := buildBackendTab(sList)
	buildGroupTab(sgList, backendTab)

	return backendTab
}

func buildBackendTab(sList []a10go.A10Server) map[string]*backend {
	backendTab := map[string]*backend{} // backendName => backend

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

	return backendTab
}

func addMember(bsg backendServiceGroup, memberName, memberPort string) backendServiceGroup {
	for _, sgm := range bsg.Members {
		if sgm.Name == memberName && sgm.Port == memberPort {
			// member found - nothing to do
			return bsg
		}
	}
	// member not found - append
	bsgm := backendSGMember{Name: memberName, Port: memberPort}
	bsg.Members = append(bsg.Members, bsgm)
	return bsg
}

func addGroupMember(b *backend, groupName, groupProtocol, memberName, memberPort string) {

	for i, bsg := range b.ServiceGroups {
		if bsg.Name == groupName {
			// group found - replace
			b.ServiceGroups[i] = addMember(b.ServiceGroups[i], memberName, memberPort)
			return
		}
	}

	// group not found - append new
	bsg := backendServiceGroup{Name: groupName, Protocol: groupProtocol}
	bsgm := backendSGMember{Name: memberName, Port: memberPort}
	bsg.Members = append(bsg.Members, bsgm)
	b.ServiceGroups = append(b.ServiceGroups, bsg)
}

func buildGroupTab(sgList []a10go.A10ServiceGroup, backendTab map[string]*backend) map[string]a10go.A10ServiceGroup {
	groupTab := map[string]a10go.A10ServiceGroup{} // groupName => group

	// scan service group table
	// this loop IS able to find all service groups (including those ones detached from virtual servers)
	for _, sg := range sgList {

		groupTab[sg.Name] = sg // build group tab

		for _, sgm := range sg.Members {
			b, found := backendTab[sgm.Name]
			if !found {
				continue // backend not found - skip
			}

			addGroupMember(b, sg.Name, sg.Protocol, sgm.Name, sgm.Port)
		}
	}

	return groupTab
}
