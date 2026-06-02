package core

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

var errEvalPingInvalidArgs = fmt.Errorf("ERR wrong number of arguments for 'ping' command")
var errSyntaxError = fmt.Errorf("ERR syntax error")
var errParseError = fmt.Errorf("ERR value is not an integer or out of range")

func invalidArgsError(cmd string) error {
	return fmt.Errorf("ERR wrong number of arguments for '%s' command", cmd)
}

var respNil = []byte("$-1\r\n")
var respOk = []byte("+OK\r\n")

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

// SET k v
func evalSETAndRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) <= 1 {
		return invalidArgsError("set")
	}

	var expirationMs int64 = -1
	key, value := args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			// if there is no expiry given
			if i == len(args) {
				return errSyntaxError
			}

			duration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errParseError
			}

			// in milliSeconds so
			expirationMs = duration * 1000
		default:
			return errSyntaxError
		}
	}

	Put(key, NewObject(value, expirationMs))
	conn.Write(respOk)
	return nil
}

// GET k
func evalGETandRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) != 1 {
		return invalidArgsError("get")
	}

	key := args[0]
	obj := Get(key)
	fmt.Println(key, obj)

	// not exist
	if obj == nil {
		conn.Write(respNil)
		return nil
	}

	// expired
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		conn.Write(respNil)
		return nil
	}

	conn.Write(Encode(obj.Value, false))
	return nil
}

func evalTTLandRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) != 1 {
		return invalidArgsError("ttl")
	}

	key := args[0]
	obj := Get(key)

	// key does not exist
	if obj == nil {
		conn.Write([]byte(":-2\r\n"))
		return nil
	}

	// key exist but no expiry
	if obj.ExpiresAt == -1 {
		conn.Write([]byte(":-1\r\n"))
		return nil
	}

	durationMs := obj.ExpiresAt - time.Now().UnixMilli()
	if durationMs < 0 {
		conn.Write([]byte(":-2\r\n"))
		return nil
	}

	conn.Write([]byte(Encode(durationMs/1000, false)))
	return nil
}

func evalDELETEAndRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) < 1 {
		return invalidArgsError("delete")
	}

	keysDeleted := 0

	for i := range args {
		if Delete(args[i]) {
			keysDeleted++
		}
	}

	conn.Write(Encode(keysDeleted, false))
	return nil
}

func evalEXPIREAndRespond(args []string, conn io.ReadWriteCloser) error {
	if len(args) != 2 {
		return invalidArgsError("expire")
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		conn.Write(Encode(0, false))
		return nil
	}

	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errParseError
	}

	expirationMs := time.Now().UnixMilli() + duration*1000
	obj.ExpiresAt = expirationMs
	conn.Write(Encode(1, false))

	return nil
}

func evalUnknownAndRespond(cmd string, args []string) error {
	return fmt.Errorf("ERR unknown command '%s', with args beginning with: '%s'", cmd, args[0])
}

func EvalAndRespond(cmd RedisCmd, conn io.ReadWriteCloser) error {
	// we are doing simple string for now
	switch cmd.Cmd {
	case "PING":
		return evalPingAndRespond(cmd.Args, conn)
	case "SET":
		return evalSETAndRespond(cmd.Args, conn)
	case "GET":
		return evalGETandRespond(cmd.Args, conn)
	case "TTL":
		return evalTTLandRespond(cmd.Args, conn)
	case "DEL":
		return evalDELETEAndRespond(cmd.Args, conn)
	case "EXPIRE":
		return evalEXPIREAndRespond(cmd.Args, conn)
	default:
		return evalUnknownAndRespond(cmd.Cmd, cmd.Args)
	}
}
