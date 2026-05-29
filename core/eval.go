package core

import (
	"fmt"
	"net"
)

var errEvalPingInvalidArgs = fmt.Errorf("ERR wrong number of arguments for 'ping' command")

func evalPingAndRespond(args []string, conn net.Conn) error {
	if len(args) >= 2 {
		return errEvalPingInvalidArgs
	}

	var err error
	if len(args) == 0 {
		_, err = conn.Write([]byte(Encode("PONG", true)))
	} else {
		_, err = conn.Write([]byte(Encode(args[0], false)))
	}

	return err
}

func EvalAndRespond(cmd RedisCmd, conn net.Conn) error {
	// we are doing simple string for now
	switch cmd.Cmd {
	case "PING":
		return evalPingAndRespond(cmd.Args, conn)
	default:
		return evalPingAndRespond(cmd.Args, conn)
	}
}
