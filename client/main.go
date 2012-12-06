package main

import (
	"discovery"
	"fmt"
)

func main() {
	var client discovery.Client
	err := client.Connect("localhost", 9000)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	fmt.Println(client.Join(&discovery.JoinRequest{Group: "group", Port: 9000}))
}
