package main

import (
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

	// 1. delete virtual servers
	putVServersDelete(c, vServersDelete)
	// 2. delete service groups
	putGroupsDelete(c, groupsDelete)
	// 3. delete servers
	putServersDelete(c, serversDelete)
	// 4. update servers
	putServersUpdate(c, serversUpdate, newList)
	// 5. create servers
	putServersCreate(c, serversCreate, newList)
	// 6. update service groups  - after 5
	putGroupsUpdate(c, groupsUpdate, newList)
	// 7. create service groups  - after 5
	putGroupsCreate(c, groupsCreate, newList)
	// 8. update virtual servers - after 7
	putVServersUpdate(c, vServersUpdate, newList)
	// 9. create virtual servers - after 7
	putVServersCreate(c, vServersCreate, newList)
}

func putVServersDelete(c *a10go.Client, serversDelete []string) {
}

func putGroupsDelete(c *a10go.Client, groupsDelete []string) {
}

func putServersDelete(c *a10go.Client, serversDelete []string) {
}

func putServersUpdate(c *a10go.Client, serversUpdate []string, newList []virtual) {
}

func putServersCreate(c *a10go.Client, serversCreate []string, newList []virtual) {
}

func putGroupsUpdate(c *a10go.Client, groupsUpdate []string, newList []virtual) {
}

func putGroupsCreate(c *a10go.Client, groupsCreate []string, newList []virtual) {
}

func putVServersUpdate(c *a10go.Client, vServersUpdate []string, newList []virtual) {
}

func putVServersCreate(c *a10go.Client, vServersCreate []string, newList []virtual) {
}
