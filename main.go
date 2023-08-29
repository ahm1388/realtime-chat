package main

import (
	"flag"
	"log"

	"github.com/ahm1388/realtime-chat/config"
	"github.com/ahm1388/realtime-chat/server"
)
 
func setupFlags () {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the server")
	flag.IntVar(&config.Port, "port", 7007, "port for the server")
}
func main () {
	setupFlags()
	log.Println("Chat server started. Listening...")
	server.RunSyncTCPServer()
}
