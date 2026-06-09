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

var wrongEncoding = fmt.Errorf("the operation is not permitted on this encoding")
var wrongType = fmt.Errorf("the operation is not permitted on this type")

func assertEncoding(val uint8, actual uint8) error {
	if getEncoding(val) != actual {
		return wrongEncoding
	}

	return nil
}

func assertType(val uint8, actual uint8) error {
	if getType(val) != actual {
		return wrongType
	}

	return nil
}
