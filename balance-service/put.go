package main

import (
	"fmt"
	"log"

	"github.com/udhos/a10-go-rest-client/a10go"
)

func put(c *a10go.Client, oldList, newList []virtual) {

	me := "put"

	// list
	serversOld, groupsOld, vServersOld := listNames(oldList)
	serversNew, groupsNew, vServersNew := listNames(newList)

	// servers
	log.Printf(me+": servers - existing: %v", serversOld)
	log.Printf(me+": servers - new: %v", serversNew)
	serversDelete, serversCreate, serversUpdate := compareSets(serversOld, serversNew)
	log.Printf(me+": servers - delete: %v", serversDelete)
	log.Printf(me+": servers - create: %v", serversCreate)
	log.Printf(me+": servers - update: %v", serversUpdate)

	// service groups
	log.Printf(me+": groups - existing: %v", groupsOld)
	log.Printf(me+": groups - new: %v", groupsNew)
	groupsDelete, groupsCreate, groupsUpdate := compareSets(groupsOld, groupsNew)
	log.Printf(me+": groups - delete: %v", groupsDelete)
	log.Printf(me+": groups - create: %v", groupsCreate)
	log.Printf(me+": groups - update: %v", groupsUpdate)

	// virtual servers
	log.Printf(me+": vServers - existing: %v", vServersOld)
	log.Printf(me+": vServers - new: %v", vServersNew)
	vServersDelete, vServersCreate, vServersUpdate := compareSets(vServersOld, vServersNew)
	log.Printf(me+": vServers - delete: %v", vServersDelete)
	log.Printf(me+": vServers - create: %v", vServersCreate)
	log.Printf(me+": vServers - update: %v", vServersUpdate)

	debug := true

	// 1. delete virtual servers
	putVServersDelete(debug, c, vServersDelete)
	// 2. delete service groups
	putGroupsDelete(debug, c, groupsDelete)
	// 3. delete servers
	putServersDelete(debug, c, serversDelete)
	// 4. update servers
	putServersUpdate(debug, c, serversUpdate, newList)
	// 5. create servers
	putServersCreate(debug, c, serversCreate, newList)
	// 6. update service groups  - after 5
	putGroupsUpdate(debug, c, groupsUpdate, newList)
	// 7. create service groups  - after 5
	putGroupsCreate(debug, c, groupsCreate, newList)
	// 8. update virtual servers - after 7
	putVServersUpdate(debug, c, vServersUpdate, newList)
	// 9. create virtual servers - after 7
	putVServersCreate(debug, c, vServersCreate, newList)
}

func putVServersDelete(debug bool, c *a10go.Client, names []string) {
	for _, s := range names {
		if debug {
			log.Printf("putVServersDelete: %s", s)
		}
		if err := c.VirtualServerDelete(s); err != nil {
			log.Printf("putVServersDelete: %s: %v", s, err)
		}
	}
}

func putGroupsDelete(debug bool, c *a10go.Client, names []string) {
	for _, s := range names {
		if debug {
			log.Printf("putGroupsDelete: %s", s)
		}
		if err := c.ServiceGroupDelete(s); err != nil {
			log.Printf("putGroupsDelete: %s: %v", s, err)
		}
	}
}

func putServersDelete(debug bool, c *a10go.Client, names []string) {
	for _, s := range names {
		if debug {
			log.Printf("putServersDelete: %s", s)
		}
		if err := c.ServerDelete(s); err != nil {
			log.Printf("putServersDelete: %s: %v", s, err)
		}
	}
}

func putServersUpdate(debug bool, c *a10go.Client, names []string, newList []virtual) {
	serversCreateUpdate("putServersUpdate", c.ServerUpdate, debug, c, names, newList)
}

func putServersCreate(debug bool, c *a10go.Client, names []string, newList []virtual) {
	serversCreateUpdate("putServersCreate", c.ServerCreate, debug, c, names, newList)
}

func serversCreateUpdate(label string, call func(string, string, []string) error, debug bool, c *a10go.Client, names []string, newList []virtual) {
	for _, s := range names {
		if debug {
			log.Printf("%s: %s", label, s)
		}
		host := findServer(newList, s)
		if host.Name == "" {
			log.Printf("%s: %s: not found", label, s)
			continue
		}
		var portList []string
		for _, p := range host.Ports {
			portList = append(portList, fmt.Sprintf("%s,%s", p.Port, p.Protocol))
		}
		if err := call(host.Name, host.Address, portList); err != nil {
			log.Printf("%s: %s: %v", label, host.Name, err)
		}
	}
}

func putGroupsUpdate(debug bool, c *a10go.Client, names []string, newList []virtual) {
	groupsCreateUpdate("putGroupsUpdate", c.ServiceGroupUpdate, debug, c, names, newList)
}

func putGroupsCreate(debug bool, c *a10go.Client, names []string, newList []virtual) {
	groupsCreateUpdate("putGroupsCreate", c.ServiceGroupCreate, debug, c, names, newList)
}

func groupsCreateUpdate(label string, call func(string, string, []string) error, debug bool, c *a10go.Client, names []string, newList []virtual) {
	for _, s := range names {
		if debug {
			log.Printf("%s: %s", label, s)
		}
		p := findPool(newList, s)
		if p.Name == "" {
			log.Printf("%s: %s: not found", label, s)
			continue
		}
		var portList []string // port = "serverName,portNumber,portProtocol"
		for _, member := range p.Members {
			for _, mp := range member.Ports {
				portList = append(portList, fmt.Sprintf("%s,%s,%s", member.Name, mp.Port, mp.Protocol))
			}
		}
		if err := call(p.Name, p.Protocol, portList); err != nil {
			log.Printf("%s: %s: %v", label, p.Name, err)
		}
	}
}

func putVServersUpdate(debug bool, c *a10go.Client, names []string, newList []virtual) {
}

func putVServersCreate(debug bool, c *a10go.Client, names []string, newList []virtual) {
}
