package resp

import (
	"bytes"
	"fmt"
	"strconv"
)

var RESP_NIL []byte = []byte("-1\r\n")

func encodeString(v string) []byte {
	var b bytes.Buffer

	b.WriteString("$")
	b.WriteString(strconv.Itoa(len(v)))
	b.WriteString("\r\n")
	b.WriteString(v)
	b.WriteString("\r\n")

	return b.Bytes()
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {

	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}

		return encodeString(v)

	case int, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))

	case []string:
		var buf bytes.Buffer

		buf.WriteString("*")
		buf.WriteString(strconv.Itoa(len(v)))
		buf.WriteString("\r\n")

		for _, item := range v {
			buf.Write(encodeString(item))
		}

		return buf.Bytes()

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