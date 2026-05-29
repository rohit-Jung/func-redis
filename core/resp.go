// Package core: things related with Redis Serialization Protocol
package core

import (
	"bytes"
	"fmt"
	"strconv"
)

func readSimpleString(data []byte) (string, int, error) {
	pos := 1

	for ; data[pos] != '\r'; pos++ {
	}

	// pos +2 for \r\n
	return string(data[1:pos]), pos + 2, nil
}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

func readInt(data []byte) (int, int, error) {
	pos := 1
	indexOfCrlf := bytes.Index(data, []byte("\r\n"))
	if indexOfCrlf == -1 {
		return 0, 0, fmt.Errorf("ERR Error Parsing")
	}

	num, err := strconv.Atoi(string(data[pos:indexOfCrlf]))
	if err != nil {
		return 0, 0, err
	}

	return num, pos + indexOfCrlf + 2, nil
}

func readBulkString(data []byte) (string, int, error) {
	pos := 1

	numBytes, err := strconv.Atoi(string(data[pos]))
	if err != nil {
		return "", 0, err
	}

	pos += 3 // initial <num>/r/n
	fmt.Println("pos", pos, pos+numBytes)
	contents := data[pos : pos+numBytes]
	pos += numBytes
	fmt.Println("numBytes", numBytes, "data", string(contents), pos)

	return string(contents), pos + int(numBytes) + 2, nil
}

func readArray(data []byte) ([]string, int, error) {
	pos := 1 // it's the inital '*'

	numCmds, err := strconv.Atoi(string(data[pos])) // number of cmds to parse
	if err != nil {
		return nil, 0, err
	}

	cmds := make([]string, numCmds)
	pos += 2

	for range numCmds {
		val, delta, err := DecodeOne(data[pos:]) // +2 is for \r\n
		if err != nil {
			return nil, 0, err
		}

		cmds = append(cmds, val.(string))
		pos += delta
	}

	return cmds, pos, nil
}

func DecodeOne(data []byte) (any, int, error) {
	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case '$':
		return readBulkString(data)
	case ':':
		return readInt(data)
	}
	return nil, 0, nil
}

func Decode(data []byte) (any, error) {
	value, _, err := DecodeOne(data)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// DecodeArrayString receive byte of data, decode em, convert to tokens  and returns
func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	// type interface ?
	ts := value.([]interface{})
	tokens := make([]string, len(ts))
	for i := range tokens {
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}

func Encode(value any, isSimple bool) string {
	switch value.(type) {
	case string:
		if isSimple {
			return fmt.Sprintf("+%s\r\n", value)
		}
		return fmt.Sprintf("$%s\r\n", value)
	default:
		// TODO: fix this
		return "+OK\r\n"
	}
}
