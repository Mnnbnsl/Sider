package store

import (
	"errors"
	"strconv"
)

func getType(te uint8) uint8 {
	return te & 0b11110000
}

func getEncoding(te uint8) uint8 {
	return te & 0b00001111
}

func AssertType(te uint8, t uint8) error {
	if getType(te) != t {
		return errors.New("The operation is not permitted on this type.")
	}
	return nil 
}

func AssertEncoding(te uint8, e uint8) error {
	if getEncoding(te) != e {
		return errors.New("The operation is not permitted on this type.")
	}
	return nil
}

func DeduceTypeEncoding(v string) (uint8, uint8) {
	oType := OBJ_TYPE_STRING

	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, OBJ_ENCODING_INT
	}

	if len(v) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR
	}

	return oType, OBJ_ENCODING_RAW
}
