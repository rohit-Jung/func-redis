package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

var errEvalPingInvalidArgs = "ERR wrong number of arguments for 'ping' command"
var errSyntaxError = "ERR syntax error"
var errParseError = "ERR value is not an integer or out of range"

// initalize valid txn commands
var txnCommands map[string]bool

func init() {
	txnCommands = map[string]bool{"EXEC": true, "DISCARD": true}
}

func invalidArgsError(cmd string) string {
	return fmt.Sprintf("ERR wrong number of arguments for '%s' command", cmd)
}

var respNil = []byte("$-1\r\n")
var respOk = []byte("+OK\r\n")
var respZero = []byte(":0\r\n")
var respOne = []byte(":1\r\n")
var respMinusTwo = []byte(":-2\r\n")
var respMinusOne = []byte(":-1\r\n")
var respQueued = []byte("+QUEUED\r\n")

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
		case "EX", "ex", "Ex", "eX":
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
		return respMinusTwo
	}

	// key exist but no expiry
	exp, isExpSet := getExpiry(obj)
	if !isExpSet {
		return respMinusOne
	}

	// key is expired; so return -2
	if exp < uint32(time.Now().UnixMilli()) {
		return respMinusTwo
	}

	// return the remaining time to expire
	durationMs := exp - uint32(time.Now().UnixMilli())
	return Encode(int64(durationMs/1000), false)
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
		return respZero
	}

	durationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return respOne
	}

	// fault: we did not add in the current time it would be done by setExpiry
	setExpiry(obj, uint32(durationSec*1000))
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

// just to test the graceful shutdown
func evalSleep(args []string) []byte {
	if len(args) != 1 {
		return Encode(invalidArgsError("sleep"), true)
	}

	duration, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return Encode(errors.New("ERROR value is not an integer or out of range"), false)
	}

	time.Sleep(time.Duration(duration) * time.Second)
	return respOk
}

func evalMULTI(c *Client) []byte {
	c.TxnBegin()
	return respOk
}

func evaluateCommand(cmd *RedisCmd, c *Client) []byte {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args)
	case "SET":
		return evalSET(cmd.Args)
	case "GET":
		return evalGET(cmd.Args)
	case "TTL":
		return evalTTL(cmd.Args)
	case "DEL":
		return evalDELETE(cmd.Args)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args)
	case "COMMAND":
		return evalPing(cmd.Args)
	case "INFO":
		return evalINFO(cmd.Args)
	case "INCR":
		return evalINCR(cmd.Args)
	case "SLEEP":
		return evalSleep(cmd.Args)
	case "CLIENT":
		return evalClient()
	case "LATENCY":
		return evalLatency()
	case "BGREWRITEAOF":
		return evalRewriteAOF()
	case "MULTI":
		return evalMULTI(c)
	case "EXEC":
		if !c.isTxn {
			return []byte("ERR EXEC used without MULTI")
		}

		return c.TxnExec()
	case "DISCARD":
		if !c.isTxn {
			return []byte("ERR DISCARD used without MULTI")
		}

		c.TxnDiscard()
		return respOk
	default:
		return evalUnknown(cmd.Cmd, cmd.Args)
	}
}

func evaluateCommandToBuffer(cmd *RedisCmd, buf *bytes.Buffer, c *Client) {
	buf.Write(evaluateCommand(cmd, c))
}

func EvalCommand(cmds RedisCmds, c io.ReadWriteCloser) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)

	for _, cmd := range cmds {
		// if it's not txns or the type assertion fails
		// means it's a plain connection
		client, ok := c.(*Client)
		if !ok {
			evaluateCommandToBuffer(cmd, buffer, client)
			continue
		}

		if !client.isTxn {
			evaluateCommandToBuffer(cmd, buffer, client)
			continue
		}

		// now we are sure it's a transaction so do we execute
		// or do we enque depends on if the command is either EXEC or DISCARD
		if !txnCommands[cmd.Cmd] {
			// if the command is not EXEC or DISCARD and it's transaction then enque
			// and respond withe queued
			client.TxnQueue(cmd)
			buffer.Write(respQueued)
		} else {
			// it's EXEC or DISCARD evaluate accordingly
			evaluateCommandToBuffer(cmd, buffer, client)
		}

	}

	return buffer.Bytes(), nil
}
