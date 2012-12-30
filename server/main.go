package main

import (
	"discovery"
	"flag"
	"fmt"
)

var port = flag.Int("port", int(discovery.DefaultPort), "Port to listen on.")

func main() {
	flag.Parse()
	server := discovery.NewServer()
	err := server.Serve(uint16(*port))
	fmt.Println("Error running server", err)
}
