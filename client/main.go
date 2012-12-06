package main

import (
	"discovery"
	"flag"
	"fmt"
	"strconv"
)

var host = flag.String("host", "localhost", "Discovery service host.")
var port = flag.Int(
	"port",
	int(discovery.DefaultPort),
	"Discovery service port.")

func main() {
	flag.Parse()
	var client discovery.Client
	err := client.Connect(*host, uint16(*port))
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Invalid command")
		return
	}

	switch args[0] {
	case "join":
		if len(args) < 4 {
			fmt.Println("client join requires <group> <host> <port>")
			return
		}
		port, err := strconv.ParseInt(args[3], 10, 16)
		if err != nil {
			fmt.Println("Invalid port:", args[3])
			return
		}
		err = client.Join(&discovery.JoinRequest{
			Group: args[1], Host: args[2], Port: uint16(port)})
	default:
	}
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Ctrl-C to exit...")
	sem := make(chan int)
	<-sem
}
