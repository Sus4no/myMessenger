package main

import "log"

const (
	PORT = ":8080"
)

func main() {
	server, err := GetServer(PORT)
	if err != nil {
		log.Print(err)
	}
	server.RunServer()
}
