package main

type virtual struct {
	Name    string
	Address string
	Port    string
	Pools   []pool
}

type pool struct {
	Name    string
	Members []server
}

type server struct {
	Name     string
	Address  string
	Port     string
	Protocol string
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
