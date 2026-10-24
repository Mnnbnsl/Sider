package resp

var (
	// RESP_NIL is the null bulk string reply.
	RESP_NIL []byte = []byte("$-1\r\n")
	// RESP_OK is the simple string reply for successful commands.
	RESP_OK []byte = []byte("+OK\r\n")
	// RESP_PONG is the reply returned by PING without arguments.
	RESP_PONG []byte = []byte("+PONG\r\n")
	// RESP_EMPTY_ARRAY is an empty array reply for unsupported discovery commands.
	RESP_EMPTY_ARRAY []byte = []byte("*0\r\n")
	// RESP_TTL_NOEXPIRE means the key exists but has no expiration set.
	RESP_TTL_NOEXPIRE []byte = []byte(":-1\r\n")
	// RESP_TTL_NOEXIST means the key does not exist or already expired.
	RESP_TTL_NOEXIST []byte = []byte(":-2\r\n")
)
