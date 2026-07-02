package core

import "fmt"

// most significant 4 bits
func getType(val uint8) uint8 {
	return val & 0b1111000
}

// least significant 4 bits
func getEncoding(val uint8) uint8 {
	return (val << 4) >> 4
}

var errWrongEncoding = fmt.Errorf("the operation is not permitted on this encoding")
var errWrongType = fmt.Errorf("the operation is not permitted on this type")

func assertEncoding(val uint8, actual uint8) error {
	if getEncoding(val) != actual {
		return errWrongEncoding
	}

	return nil
}

func assertType(val uint8, actual uint8) error {
	if getType(val) != actual {
		return errWrongType
	}

	return nil
}
