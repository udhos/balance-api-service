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
	Name    string
	Address string
	Port    string
}
