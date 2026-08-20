package resp

import "fmt"

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {

	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}

		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))

	case int:
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
		return []byte("$-1\r\n")
	}

	return []byte{}
}