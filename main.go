package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(res http.ResponseWriter, req *http.Request)
}

type SimpleServer struct {
	address string
	proxy   *httputil.ReverseProxy
}

type LoadBalancerService struct {
	port            string
	roundRobinCount int
	servers         []Server
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func newServer(address string) *SimpleServer {
	serverUrl, err := url.Parse(address)
	handleError(err)

	return &SimpleServer{
		address: address,
		proxy:   httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func LoadBalancer(port string, servers []Server) *LoadBalancerService {
	return &LoadBalancerService{
		port:            port,
		roundRobinCount: 0,
		servers:         servers,
	}
}

func (m *SimpleServer) Address() string {
	return m.address
}

func (m *SimpleServer) IsAlive() bool {
	return true
}

func (m *SimpleServer) Serve(res http.ResponseWriter, req *http.Request) {
	m.proxy.ServeHTTP(res, req)
}

func (b *LoadBalancerService) GetNextServer() Server {
	server := b.servers[b.roundRobinCount%len(b.servers)]

	for !server.IsAlive() {
		b.roundRobinCount++
		server = b.servers[b.roundRobinCount%len(b.servers)]
	}
	b.roundRobinCount++
	return server
}

func (b *LoadBalancerService) ServeProxy(res http.ResponseWriter, req *http.Request) {
	destinationServer := b.GetNextServer()
	fmt.Printf("Redirecting to %s\n", destinationServer.Address())
	destinationServer.Serve(res, req)

}

func main() {
	servers := []Server{
		newServer("https://thenetnaija.me"),
		newServer("https://google.com"),
		newServer("https://meta.com"),
	}

	b := LoadBalancer(":8000", servers)

	handleRedirect := func(res http.ResponseWriter, req *http.Request) {
		b.ServeProxy(res, req)
	}

	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Load Balancer listening on port %s\n", b.port)
	err := http.ListenAndServe(b.port, nil)
	handleError(err)
}
