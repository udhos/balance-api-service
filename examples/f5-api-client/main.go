package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/e-XpertSolutions/f5-rest-client/f5/net"
)

func sexyPrint(label string, a interface{}) {
	j, err := json.MarshalIndent(a, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("DEBUG ", label, "\n", string(j), "\n")
}

func main() {

	f5Host := "https://10.255.255.120"

	// 1) Basic Authentication
	f5Client, err := f5.NewBasicClient(f5Host, "admin", "admin")

	// 2) Token Based Authentication
	// f5Client, err := f5.NewTokenClient(f5Host, "admin", "admin", "tmos", true)

	if err != nil {
		log.Fatal(err)
	}
	f5Client.DisableCertCheck()

	netList(f5Client)

	vsList(f5Client)
}

func netList(f5Client *f5.Client) {
	netClient := net.New(*f5Client) // client for net
	self, err := netClient.Self().ListAll()
	if err != nil {
		log.Fatal(err)
	}
	sexyPrint("net SelfIP List:", self)
}

func vsList(f5Client *f5.Client) {
	ltmClient := ltm.New(*f5Client) // client for ltm api

	// query the /ltm/virtual API
	vsConfigList, err := ltmClient.Virtual().ListAll()
	if err != nil {
		log.Fatal(err)
	}
	sexyPrint("ltm virtual List:", vsConfigList)
}
