// Package server: server related funcs
package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/rohit-Jung/func-redis/config"
	"github.com/rohit-Jung/func-redis/core"
)

func RunSyncServer() error {
	log.Println("starting sync tcp server on", config.Host, config.Port)

	connClients := 1

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	for {
		conn, err := lsnr.Accept() // this is a blocking call
		if err != nil {
			log.Printf("Error accepting connection %v", err)
			continue
		}

		log.Println("Connected to ", conn.RemoteAddr())
		connClients += 1

		// breaks only when client disconnects or QUIT command is issued
		// (so sync server is problem here)
		for {
			cmds, err := core.ReadCommands(conn) // read the command into string

			if err != nil {
				conn.Close()
				connClients -= 1
				log.Println("Closing the connection")

				if errors.Is(err, io.EOF) {
					break
				}

				break
			}

			cmds.Respond(conn)
		}
	}
}
