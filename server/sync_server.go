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

func readCommand(conn net.Conn) (*core.RedisCmd, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		log.Print("oops reading error", err)
		return nil, err
	}

	tokens, err := core.DecodeArrayString(buf[:n])
	if err != nil {
		log.Print("Error decoding to array of strings")
	}

	fmt.Println("array strings", tokens)
	return &core.RedisCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}, nil
}

func respondError(err error, conn net.Conn) {
	respErrFormat := fmt.Sprintf("-%s\r\n", err)
	conn.Write([]byte(respErrFormat))
}

func respondCmd(cmd *core.RedisCmd, conn net.Conn) {
	err := core.EvalAndRespond(*cmd, conn)
	if err != nil {
		log.Print("Encountered write error", err)
		respondError(err, conn)
	}
}

func RunSyncServer() {
	log.Println("starting sync tcp server on", config.Host, config.Port)

	connClients := 1

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	lsnr, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Error initiating tcp connection", err)
	}

	for {
		conn, err := lsnr.Accept() // this is a blocking call
		if err != nil {
			log.Printf("Error accepting connection %v", err)
			continue
		}

		log.Println("Connected to ", conn.RemoteAddr())
		connClients += 1

		for {
			cmd, err := readCommand(conn) // read the command into string
			if err != nil {
				conn.Close()
				connClients -= 1
				log.Println("Closing the connection")

				if errors.Is(err, io.EOF) {
					break
				}

				break
			}

			respondCmd(cmd, conn)
		}
	}
}
