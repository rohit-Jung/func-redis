package core

import (
	"fmt"
	"io"
	"log"
)

type RedisCmd struct {
	Cmd  string
	Args []string
}

func ReadCommand(conn io.ReadWriteCloser) (*RedisCmd, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		log.Print("oops reading error", err)
		return nil, err
	}

	tokens, err := DecodeArrayString(buf[:n])
	if err != nil {
		log.Print("Error decoding to array of strings")
	}

	return &RedisCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}, nil
}

func respondError(err error, conn io.ReadWriteCloser) {
	respErrFormat := fmt.Sprintf("-%s\r\n", err)
	conn.Write([]byte(respErrFormat))
}

func (cmd RedisCmd) Respond(conn io.ReadWriteCloser) {
	err := EvalAndRespond(cmd, conn)
	if err != nil {
		log.Print("Encountered write error\t", err)
		respondError(err, conn)
	}
}
