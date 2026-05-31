package core

import (
	"fmt"
	"io"
)

var errEvalPingInvalidArgs = fmt.Errorf("ERR wrong number of arguments for 'ping' command")

func evalPingAndRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) >= 2 {
		return errEvalPingInvalidArgs
	}

	var err error
	if len(args) == 0 {
		_, err = conn.Write(Encode("PONG", true))
	} else {
		_, err = conn.Write(Encode(args[0], false))
	}

	return err
}

func EvalAndRespond(cmd RedisCmd, conn io.ReadWriteCloser) error {
	// we are doing simple string for now
	switch cmd.Cmd {
	case "PING":
		return evalPingAndRespond(cmd.Args, conn)
	default:
		return evalPingAndRespond(cmd.Args, conn)
	}
}
