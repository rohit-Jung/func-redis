package core

type Obj struct {
	TypeEncoding uint8
	Value        any

	// ExpiresAt    int64

	// we are now using the expiry dictionary
	// also we are setting the 32bit accessed at while the redis implementation has 24 bits
	// this is done as golang doesn't have bitfields like c
	// and it helps keep things simple for now
	lastAccessedAt uint32
}

// first 4 bytes are type and rest 4 are encodings
var OBJ_TYPE_STRING uint8 = 0 << 4

var OBJ_ENCODING_RAW uint8 = 0
var OBJ_ENCODING_INT uint8 = 1
var OBJ_ENCODING_EMBSTR uint8 = 8
