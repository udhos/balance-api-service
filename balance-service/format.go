package main

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

// serverNames extracts server names from virtual list
func serverNames(vsList []virtual) []string {
	var servers []string

	for _, vs := range vsList {
		for _, p := range vs.Pools {
			for _, m := range p.Members {
				servers = append(servers, m.Name)
			}
		}
	}

	return servers
}
