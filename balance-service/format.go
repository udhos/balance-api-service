package main

import (
	"log"
)

type virtual struct {
	Name    string
	Address string // listen addr
	Pools   []pool
}

type pool struct {
	Port     string   // listen port
	Protocol string   // listen proto
	Name     string   // pool name
	Members  []server // backend servers
}

type server struct {
	Name    string
	Address string       // backend server address
	Ports   []serverPort // backend server port/proto
}

type serverPort struct {
	Port     string // backend server port
	Protocol string // backend server proto
}

func A10ProtocolName(number string) string {
	protoName := "unknown"
	switch number {
	case "2":
		protoName = "tcp"
	case "3":
		protoName = "udp"
	default:
		log.Printf("A10ProtocolName: error: [%s]", number)
		protoName = "unknown:" + number
	}
	return protoName
}

func A10ProtocolNumber(name string) string {
	number := "unknown"
	switch name {
	case "tcp":
		number = "2"
	case "udp":
		number = "3"
	default:
		log.Printf("A10ProtocolNumber: error: [%s]", name)
		number = "unknown:" + name
	}
	return number
}

// listNames extracts all names from virtual list
func listNames(vsList []virtual) (servers, groups, vServers []string) {

	for _, vs := range vsList {
		vServers = append(vServers, vs.Name)
		for _, p := range vs.Pools {
			groups = append(groups, p.Name)
			for _, m := range p.Members {
				servers = append(servers, m.Name)
			}
		}
	}

	return
}

func findServer(vsList []virtual, name string) server {
	for _, vs := range vsList {
		for _, p := range vs.Pools {
			for _, m := range p.Members {
				if m.Name == name {
					return m
				}
			}
		}
	}

	return server{}
}

func findPool(vsList []virtual, name string) pool {
	for _, vs := range vsList {
		for _, p := range vs.Pools {
			if p.Name == name {
				return p
			}
		}
	}

	return pool{}
}

func findVirtual(vsList []virtual, name string) virtual {
	for _, vs := range vsList {
		if vs.Name == name {
			return vs
		}
	}

	return virtual{}
}
