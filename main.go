package main

import (
	"discovery"
	"flag"
	"fmt"
)

var port = flag.Int("port", 9000, "Port to listen on.")

func main() {
	flag.Parse()
	var server discovery.Server
	server.Init()
	err := server.Serve(uint16(*port))
	fmt.Println("Error running server", err)
}
