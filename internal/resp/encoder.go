package resp

import "fmt"

var RESP_NIL []byte = []byte("-1\r\n")

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {

	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}

		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))

	case int, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))

	case []interface{}:
		var result []byte

		result = append(result, fmt.Sprintf("*%d\r\n", len(v))...)

		for _, item := range v {
			result = append(result, Encode(item, false)...)
		}

		return result

	case nil:
		// Null bulk string
		return RESP_NIL
	}

	return []byte{}
}