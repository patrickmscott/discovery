package main

import (
	"discovery"
	"flag"
	"log"
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
		log.Println("Error connecting:", err)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		log.Println("Invalid command")
		return
	}

	switch args[0] {
	case "join":
		if len(args) < 4 {
			log.Println("client join requires <group> <host> <port>")
			return
		}
		var port int64
		port, err = strconv.ParseInt(args[3], 10, 16)
		if err != nil {
			log.Println("Invalid port:", args[3])
			return
		}
		err = client.Join(&discovery.ServiceDef{
			Group: args[1], Host: args[2], Port: uint16(port)})
	default:
	}
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("Ctrl-C to exit...")
	<-make(chan int)
}
