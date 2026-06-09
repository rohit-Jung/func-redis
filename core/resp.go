// Package core: things related with Redis Serialization Protocol
package core

import (
	"bytes"
	"errors"
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

// this was majorly separated from readInt not because
// of reuse but because of pos value; it had to be 1 for readInt but not for other use cases
func readNums(data []byte) (int, int, error) {
	numInBytes, _, ok := bytes.Cut(data, []byte("\r\n"))
	if !ok {
		return 0, 0, fmt.Errorf("ERR Error Parsing")
	}

	num, err := strconv.Atoi(string(numInBytes))
	if err != nil {
		return 0, 0, err
	}

	return num, len(numInBytes) + 2, nil
}

func readInt(data []byte) (int, int, error) {
	pos := 1
	num, delta, err := readNums(data[pos:])
	if err != nil {
		return 0, 0, err
	}

	return num, pos + delta, nil
}

func readBulkString(data []byte) (string, int, error) {
	pos := 1
	numBytes, delta, err := readNums(data[pos:])
	if err != nil {
		return "", 0, err
	}

	pos += delta
	contents := data[pos : pos+numBytes]
	return string(contents), pos + int(numBytes) + 2, nil // +2  for last \r\n
}

func readArray(data []byte) ([]any, int, error) {
	pos := 0

	numCmds, delta, err := readInt(data)
	if err != nil {
		return nil, 0, err
	}

	cmds := make([]any, numCmds)
	pos += delta

	for i := range numCmds {
		val, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}

		cmds[i] = val
		pos += delta
	}

	return cmds, pos, nil
}

func DecodeOne(data []byte) (any, int, error) {
	if len(data) <= 0 {
		return "", 0, fmt.Errorf("ERR error while decoding. len not sufficient")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case '$':
		return readBulkString(data)
	case ':':
		return readInt(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil
}

func Decode(data []byte) ([]any, error) {
	if len(data) <= 0 {
		return nil, errors.New("ERR: not enough length to decode")
	}

	result := make([]any, 0)
	index := 0

	// up till index doesn't crosses len(data) which will not happen
	for index < len(data) {
		value, delta, err := DecodeOne(data[index:])
		if err != nil {
			return nil, err
		}

		index += delta
		result = append(result, value)
	}

	return result, nil
}

func encodeString(v string) []byte {
	return fmt.Appendf(nil, "+%s\r\n", v)
}

func Encode(value any, isSimple bool) []byte {
	// i am dumb i realised at this point (again)
	switch v := value.(type) {
	case string:
		if isSimple {
			return encodeString(v)
		}
		return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(v), value)
	case int, int8, int16, int32, int64:
		return fmt.Appendf(nil, ":%d\r\n", v)
	case []string:
		buf := bytes.NewBuffer(nil)
		for _, val := range value.([]string) {
			buf.Write(encodeString(val))
		}

		return fmt.Appendf(nil, "*%d\r\n%s", len(v), buf.Bytes())
	default:
		return respNil
	}
}
