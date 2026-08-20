package resp

import (
	"errors"
	"strconv"
) 

// Implementing RESP2 for now.
// + simple string
// : integers
// - errors
// $ bulk string
// $-1\r\n null bulk string
// * array

var (
	ErrIncomplete = errors.New("Incomplete data")
	ErrInvalid = errors.New("Invalid data")
	ErrNoData = errors.New("No data")
)

func readSimpleString(data []byte) (string, int, error) {
	for i := 1; i < len(data); i++ {
		if data[i] == '\r' {
			// No \n means incomplete bytestream data
			if i+1 >= len(data) {
				return "", 0, ErrIncomplete
			}
			
			// if no \r\n then invalid data
			if data[i+1] != '\n' {
				return "", 0, ErrInvalid
			}
			return string(data[1:i]), i+2, nil
		}
	}
	return "", 0, ErrIncomplete
}

func readInteger(data []byte) (int, int, error) {
	if len(data) < 2 {
    	return 0, 0, ErrIncomplete
	}

	for i := 1; i < len(data); i++ {
		if data[i] >= '0' && data[i] <= '9' {
			continue
		}
		if i == 1 && data[i] == '-' {
			continue // negative int
		} 
		if data[i] != '\r' {
			return 0, 0, ErrInvalid // data like 123abc
		} else {
			if i+1 >= len(data) {
				return 0, 0, ErrIncomplete
			} else if  data[i+1] != '\n' {
				return 0, 0, ErrInvalid
			}
			num, err := strconv.Atoi(string(data[1:i]))
			if err != nil {
				return 0, 0, ErrInvalid
			}
			return num, i+2, nil
		}
	}
	return 0, 0, ErrIncomplete
}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

func readBulkString(data []byte) (string, int, error) {
	if len(data) < 3 {
		return "", 0, ErrIncomplete
	}

	// Find the \r\n after the length
	i := 1

	for ; i < len(data); i++ {
		if data[i] >= '0' && data[i] <= '9' {
			continue
		} 
		if data[i] == '-' && i == 1 {
			continue
		} 
		if data[i] != '\r' {
			return "", 0, ErrInvalid
		} 
		if i+1 >= len(data) {
			return "", 0, ErrIncomplete
		} 
		if data[i+1] != '\n' {
			return "", 0, ErrInvalid
		}

		break
	}

	if i == len(data) {
		return "", 0, ErrIncomplete
	}

	length, err := strconv.Atoi(string(data[1:i]))
	if err != nil {
		return "", 0, ErrInvalid
	}

	// Null bulk string: $-1\r\n
	if length == -1 {
		return "", i + 2, nil
	}

	// Other negative lengths are invalid.
	if length < 0 {
		return "", 0, ErrInvalid
	}

	// Start of the actual string.
	start := i + 2
	end := start + length

	// We need the string itself + \r\n.
	if end+2 > len(data) {
		return "", 0, ErrIncomplete
	}

	if data[end] != '\r' || data[end+1] != '\n' {
		return "", 0, ErrInvalid
	}

	return string(data[start:end]), end + 2, nil
}

func readArray(data []byte) (interface{}, int, error) {
	if len(data) < 3 {
		return nil, 0, ErrIncomplete
	}

	// Find the \r\n after the array length.
	i := 1

	for ; i < len(data); i++ {
		if data[i] >= '0' && data[i] <= '9' {
			continue
		}

		if data[i] == '-' && i == 1 {
			continue
		}

		if data[i] != '\r' {
			return nil, 0, ErrInvalid
		}

		if i+1 >= len(data) {
			return nil, 0, ErrIncomplete
		}

		if data[i+1] != '\n' {
			return nil, 0, ErrInvalid
		}

		break
	}

	if i == len(data) {
		return nil, 0, ErrIncomplete
	}

	count, err := strconv.Atoi(string(data[1:i]))
	if err != nil {
		return nil, 0, ErrInvalid
	}

	// Null array: *-1\r\n
	if count == -1 {
		return nil, i + 2, nil
	}

	// Other negative array lengths are invalid.
	if count < 0 {
		return nil, 0, ErrInvalid
	}

	array := make([]interface{}, 0, count)

	// Start reading elements after *<count>\r\n
	pos := i + 2

	for j := 0; j < count; j++ {
		if pos >= len(data) {
			return nil, 0, ErrIncomplete
		}

		value, consumed, err := decodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}

		array = append(array, value)
		pos += consumed
	}

	return array, pos, nil
}

func decodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("No data")
	}
	symbol := data[0]
	switch symbol {
	case '+':
		return readSimpleString(data)
	case ':':
		return readInteger(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	case '-':
		return readError(data)
	default :
		return nil, 0, ErrInvalid
	}
}

func DecodeArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	ts := value.([]interface{})
	tokens := make([]string, len(ts))
	for i := range tokens {
		tokens[i] = ts[i].(string)
	}

	return tokens, nil
}

func Decode(data []byte) (interface{}, error) {
	if len(data) == 0 {
		return nil, ErrNoData
	}

	value, _, err := decodeOne(data)
	return value, err
}

