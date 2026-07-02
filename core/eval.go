package core

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

var errEvalPingInvalidArgs = "ERR wrong number of arguments for 'ping' command"
var errSyntaxError = "ERR syntax error"
var errParseError = "ERR value is not an integer or out of range"

func invalidArgsError(cmd string) string {
	return fmt.Sprintf("ERR wrong number of arguments for '%s' command", cmd)
}

var respNil = []byte("$-1\r\n")
var respOk = []byte("+OK\r\n")
var respTwo = []byte(":-2\r\n")
var respOne = []byte(":-1\r\n")

func evalPing(args []string) []byte {
	if len(args) >= 2 {
		return Encode(errEvalPingInvalidArgs, true)
	}

	if len(args) == 0 {
		return Encode("PONG", true)
	} else {
		return Encode(args[0], true)
	}
}

// SET k v
func evalSET(args []string) []byte {
	if len(args) <= 1 {
		return Encode(invalidArgsError("set"), true)
	}

	var expirationMs int64 = -1
	key, value := args[0], args[1]
	oType, oEncoding := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			// if there is no expiry given
			if i == len(args) {
				return Encode(errSyntaxError, true)
			}

			duration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return Encode(errParseError, true)
			}

			// in milliSeconds so
			expirationMs = duration * 1000
		default:
			return Encode(errSyntaxError, true)
		}
	}

	Put(key, NewObject(value, expirationMs, oType, oEncoding))
	return respOk
}

// GET k
func evalGET(args []string) []byte {
	if len(args) != 1 {
		return Encode(invalidArgsError("get"), true)
	}

	key := args[0]
	obj := Get(key)

	// not exist
	if obj == nil {
		return respNil
	}

	// expired
	if hasExpired(obj) {
		return respNil
	}

	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {
	if len(args) != 1 {
		return Encode(invalidArgsError("ttl"), true)
	}

	key := args[0]
	obj := Get(key)

	// key does not exist
	if obj == nil {
		return respTwo
	}

	// key exist but no expiry
	exp, isExpSet := getExpiry(obj)
	if !isExpSet {
		return respOne
	}

	durationMs := exp - uint32(time.Now().UnixMilli())
	if exp < uint32(time.Now().UnixMilli()) {
		return respTwo
	}

	return []byte(Encode(durationMs/1000, false))
}

func evalDELETE(args []string) []byte {
	if len(args) < 1 {
		return Encode(invalidArgsError("delete"), true)
	}

	keysDeleted := 0

	for i := range args {
		if Delete(args[i]) {
			keysDeleted++
		}
	}

	return Encode(keysDeleted, false)
}

func evalEXPIRE(args []string) []byte {
	if len(args) != 2 {
		return Encode(invalidArgsError("expire"), true)
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		return (Encode(0, false))
	}

	duration, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errParseError, true)
	}

	expirationMs := time.Now().UnixMilli() + duration*1000
	setExpiry(obj, uint32(expirationMs))
	return Encode(1, false)
}

// if no object creates and puts
// increases it and stores the formatInt
func evalINCR(args []string) []byte {
	if len(args) < 1 {
		return Encode(invalidArgsError("incr"), true)
	}

	key := args[0]
	obj := Get(key)

	// dumb ways to die; if newObject then use that or reassign
	if obj == nil {
		obj = NewObject("0", -1, OBJ_TYPE_STRING, OBJ_ENCODING_INT)
		Put(key, obj)
	}

	// check encodings and type
	if err := assertEncoding(obj.TypeEncoding, OBJ_ENCODING_INT); err != nil {
		return Encode(err, false)
	}

	if err := assertType(obj.TypeEncoding, OBJ_TYPE_STRING); err != nil {
		return Encode(err, false)
	}

	i, _ := strconv.ParseInt(obj.Value.(string), 10, 64)
	i += 1
	obj.Value = strconv.FormatInt(i, 10)

	return Encode(i, false)
}

func evalINFO(_ []string) []byte {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString("# Keyspace\r\n")
	for i, keystat := range KeymapsStat {
		dbInfo := fmt.Sprintf("db%dkeys=%d,expires=0,avg_ttl=0\r\n", i, keystat["keys"])
		buffer.WriteString(dbInfo)
	}

	return Encode(buffer.String(), false)
}

func evalUnknown(cmd string, args []string) []byte {
	return Encode(fmt.Sprintf("ERR unknown command '%s', with args beginning with: '%s'", cmd, args[0]), true)
}

func evalRewriteAOF() []byte {
	DumpAllAof()
	return respOk
}

func evalLatency() []byte {
	return Encode([]string{}, false)
}

func evalClient() []byte {
	return respOk
}

func EvalCommand(cmds RedisCmds) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buffer.Write(evalPing(cmd.Args))
		case "SET":
			buffer.Write(evalSET(cmd.Args))
		case "GET":
			buffer.Write(evalGET(cmd.Args))
		case "TTL":
			buffer.Write(evalTTL(cmd.Args))
		case "DEL":
			buffer.Write(evalDELETE(cmd.Args))
		case "EXPIRE":
			buffer.Write(evalEXPIRE(cmd.Args))
		case "COMMAND":
			buffer.Write(evalPing(cmd.Args))
		case "INFO":
			buffer.Write(evalINFO(cmd.Args))
		case "INCR":
			buffer.Write(evalINCR(cmd.Args))
		case "CLIENT":
			buffer.Write(evalClient())
		case "LATENCY":
			buffer.Write(evalLatency())
		case "BGREWRITEAOF":
			buffer.Write(evalRewriteAOF())
		default:
			buffer.Write(evalUnknown(cmd.Cmd, cmd.Args))
		}
	}

	return buffer.Bytes(), nil
}
