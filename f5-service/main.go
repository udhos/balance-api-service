package main

import (
	"encoding/json"
	"log"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/net"
)

func sexyPrint(label string, a interface{}) {
	j, err := json.MarshalIndent(a, "", "   ")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("DEBUG ", label, ":\n", string(j))
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
	netClient := net.New(*f5Client)
	self, err := netClient.Self().ListAll()
	if err != nil {
		log.Fatal(err)
	}
	sexyPrint("SelfIP List:", self)
}
