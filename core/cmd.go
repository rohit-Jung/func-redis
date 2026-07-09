package core

import (
	"fmt"
	"io"
	"log"
	"strings"
)

type RedisCmd struct {
	Cmd  string
	Args []string
}

type RedisCmds []*RedisCmd

func toArrayString(data []any) ([]string, error) {
	result := make([]string, len(data))
	for i := range data {
		result[i] = data[i].(string)
	}
	return result, nil
}

func ReadCommands(conn io.ReadWriteCloser) (RedisCmds, error) {
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	values, err := Decode(buf[:n])
	if err != nil {
		return nil, err
	}

	cmds := make([]*RedisCmd, 0)
	for _, value := range values {
		tokens, err := toArrayString(value.([]any))

		if err != nil {
			return nil, err
		}
		cmds = append(
			cmds,
			&RedisCmd{
				Cmd:  strings.ToUpper(tokens[0]),
				Args: tokens[1:],
			},
		)
	}

	return cmds, nil
}

func respondError(err error, conn io.ReadWriteCloser) {
	respErrFormat := fmt.Sprintf("-%s\r\n", err)
	conn.Write([]byte(respErrFormat))
}

func (cmds RedisCmds) Respond(conn io.ReadWriteCloser) {
	buf, err := EvalCommand(cmds, conn)
	if err != nil {
		log.Print("Encountered write error\t", err)
		respondError(err, conn)
	}

	conn.Write(buf)
}
