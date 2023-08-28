package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/ahm1388/realtime-chat/config"
)

func readCommand(c net.Conn) (string, error) {
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes, then repeated read until
	// we get EOF or designated delimiter
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

func RunSyncTCPServer() {
	log.Println("Starting a synchronous TCP server on", config.Host, config.Port)

	var con_clients int = 0

	// listening for the configured host port
	lsnr, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
	if err != nil {
		panic(err)
	}

	for {
		// blocking call; waiting for the client to connect
		c, err := lsnr.Accept()
		if err != err {
			panic(err)
		}

		// increment the number of concurrent clients
		con_clients += 1
		log.Println("client connected with address:", c.RemoteAddr(), "concurrent clients", con_clients)
		
		justJoined := true
		var name string = ""
		for {
			if justJoined {
				if _, err := c.Write([]byte("name: ")); err != nil {
					log.Print("err write:", err)
					c.Close()
					con_clients -= 1
					break
				}

				// read the name provided by the user
				name, err = readCommand(c)
				if err != nil {
						log.Println("err reading name:", err)
						c.Close()
						con_clients -= 1
						break
				}
		
				name = strings.TrimSpace(name)
		
				if name == "" {
						if _, err := c.Write([]byte("Invalid name. Please try again.\n")); err != nil {
								log.Print("err write:", err)
						}
						continue
				}

				// Send a welcome message
				if err = respond("Welcome to the chat room, " + name + "!\n", c); err != nil {
					log.Print("err write:", err)
				}
				
				log.Println("User's name:", name)
				justJoined = false
				
			} else {
				// over the socket, continuously read the command and print it out
				cmd, err := readCommand(c)
				if strings.TrimSpace(cmd) == "exit" {
					c.Close()
					con_clients -= 1
					log.Println("client disconnected", c.RemoteAddr(), "concurrent clients", con_clients)
					break
				}
				if err != nil {
					c.Close()
					con_clients -= 1
					log.Println("client disconnected", c.RemoteAddr(), "concurrent clients", con_clients)
					if err == io.EOF {
						log.Println("err", err)
					}
					break
				}

				// prefix the message with the user's name
				cmdWithUserName := "[" + name + "] " + cmd
				log.Println("command", cmd)
				if err = respond(cmdWithUserName, c); err != nil {
					log.Print("err write:", err)
				}
			}
		}
	}
}
