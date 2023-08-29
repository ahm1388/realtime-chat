package server

import (
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/ahm1388/realtime-chat/config"
)

var mutex = &sync.Mutex{}
var clients = make(map[net.Conn]bool)
var clientNames = make(map[net.Conn]string)

func readCommand(c net.Conn) (string, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return "", err
	}
	return string(buf[:n]), nil
}

func respond(cmd string, c net.Conn) error {
	if _, err := c.Write([]byte(cmd)); err != nil {
		return err
	}
	return nil
}

func addClient(c net.Conn) {
	mutex.Lock()
	clients[c] = true
	mutex.Unlock()
}

func removeClient(c net.Conn) {
	mutex.Lock()
	delete(clients, c)
	delete(clientNames, c)
	mutex.Unlock()
}

func broadcastMessage(message string) {
	mutex.Lock()
	for client := range clients {
		client.Write([]byte(message))
	}
	mutex.Unlock()
}

func handleClient(c net.Conn, con_clients *int) {
	justJoined := true

	for {
		if justJoined {
			if _, err := c.Write([]byte("name: ")); err != nil {
				log.Print("err write:", err)
				removeClient(c)
				mutex.Lock()
				*con_clients -= 1
				mutex.Unlock()
				c.Close()
				break
			}

			name, err := readCommand(c)
			if err != nil {
				log.Println("err reading name:", err)
				removeClient(c)
				mutex.Lock()
				*con_clients -= 1
				mutex.Unlock()
				c.Close()
				break
			}

			name = strings.TrimSpace(name)
			if name == "" {
				if _, err := c.Write([]byte("Invalid name. Please try again.\n")); err != nil {
					log.Print("err write:", err)
				}
				continue
			}

			clientNames[c] = name

			welcomeMessage := "Welcome to the chat room, " + name + "!\n"
			respond(welcomeMessage, c)
			broadcastMessage(name + " has joined the chat.\n")

			log.Println("User's name:", name)
			addClient(c)
			justJoined = false

		} else {
			cmd, err := readCommand(c)

			if err != nil || strings.TrimSpace(cmd) == "exit" {
				name := clientNames[c]
				broadcastMessage(name + " has left the chat.\n")
				removeClient(c)
				mutex.Lock()
				*con_clients -= 1
				mutex.Unlock()
				c.Close()
				log.Println("client disconnected", c.RemoteAddr(), "concurrent clients", *con_clients)
				break
			}

			name := clientNames[c]
			cmdWithUserName := "[" + name + "] " + cmd
			broadcastMessage(cmdWithUserName)
		}
	}
}

func RunSyncTCPServer() {
	log.Println("Starting a synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	lsnr, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	for {
		c, err := lsnr.Accept()
		if err != nil {
			panic(err)
		}

		mutex.Lock()
		con_clients += 1
		mutex.Unlock()

		log.Println("client connected with address:", c.RemoteAddr(), "concurrent clients", con_clients)

		go handleClient(c, &con_clients)
	}
}
