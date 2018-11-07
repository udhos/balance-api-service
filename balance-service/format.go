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
