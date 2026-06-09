package core

import "strconv"

// Similar to `tryObjectEncoding` in redis
// if you can parse to string it's INT
// if length <= 44 it's embstr else raw
func deduceTypeEncoding(v string) (uint8, uint8) {
	oType := OBJ_TYPE_STRING

	if _, err := strconv.ParseInt(v, 10, 64); err != nil {
		return oType, OBJ_ENCODING_INT
	}

	if len(v) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR
	}

	return oType, OBJ_ENCODING_RAW
}
