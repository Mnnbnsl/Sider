package resp

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  error
	}{
		// --------------------------------------------------
		// Simple String
		// --------------------------------------------------
		{
			name:     "simple string",
			input:    "+OK\r\n",
			expected: "OK",
		},
		{
			name:     "empty simple string",
			input:    "+\r\n",
			expected: "",
		},

		// --------------------------------------------------
		// Error
		// --------------------------------------------------
		{
			name:     "error",
			input:    "-ERR unknown command\r\n",
			expected: "ERR unknown command",
		},
		{
			name:     "error with spaces",
			input:    "-WRONGTYPE Operation against a key\r\n",
			expected: "WRONGTYPE Operation against a key",
		},

		// --------------------------------------------------
		// Integer
		// --------------------------------------------------
		{
			name:     "positive integer",
			input:    ":100\r\n",
			expected: 100,
		},
		{
			name:     "zero integer",
			input:    ":0\r\n",
			expected: 0,
		},
		{
			name:     "negative integer",
			input:    ":-42\r\n",
			expected: -42,
		},
		{
			name:     "large integer",
			input:    ":123456789\r\n",
			expected: 123456789,
		},

		// --------------------------------------------------
		// Bulk String
		// --------------------------------------------------
		{
			name:     "bulk string",
			input:    "$5\r\nhello\r\n",
			expected: "hello",
		},
		{
			name:     "empty bulk string",
			input:    "$0\r\n\r\n",
			expected: "",
		},
		{
			name:     "bulk string with spaces",
			input:    "$11\r\nhello world\r\n",
			expected: "hello world",
		},
		{
			name:     "null bulk string",
			input:    "$-1\r\n",
			expected: "",
		},

		// --------------------------------------------------
		// Empty Array
		// --------------------------------------------------
		{
			name:     "empty array",
			input:    "*0\r\n",
			expected: []interface{}{},
		},

		// --------------------------------------------------
		// Array
		// --------------------------------------------------
		{
			name:  "single element array",
			input: "*1\r\n$4\r\nPING\r\n",
			expected: []interface{}{
				"PING",
			},
		},
		{
			name:  "multiple element array",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			expected: []interface{}{
				"SET",
				"foo",
				"bar",
			},
		},

		// --------------------------------------------------
		// Mixed Array
		// --------------------------------------------------
		{
			name:  "mixed type array",
			input: "*3\r\n+OK\r\n:42\r\n$5\r\nhello\r\n",
			expected: []interface{}{
				"OK",
				42,
				"hello",
			},
		},

		// --------------------------------------------------
		// Nested Array
		// --------------------------------------------------
		{
			name: "nested array",
			input: "*2\r\n*2\r\n:1\r\n:2\r\n*1\r\n+OK\r\n",
			expected: []interface{}{
				[]interface{}{
					1,
					2,
				},
				[]interface{}{
					"OK",
				},
			},
		},

		// --------------------------------------------------
		// Null Array
		// --------------------------------------------------
		{
			name:     "null array",
			input:    "*-1\r\n",
			expected: nil,
		},

		// --------------------------------------------------
		// Invalid Input
		// --------------------------------------------------
		{
			name:    "invalid type",
			input:   "@hello\r\n",
			wantErr: ErrInvalid,
		},
		{
			name:    "invalid integer",
			input:   ":abc\r\n",
			wantErr: ErrInvalid,
		},
		{
			name:    "invalid bulk string length",
			input:   "$abc\r\n",
			wantErr: ErrInvalid,
		},
		{
			name:    "invalid array length",
			input:   "*abc\r\n",
			wantErr: ErrInvalid,
		},

		// --------------------------------------------------
		// Incomplete Input
		// --------------------------------------------------
		{
			name:    "incomplete simple string",
			input:   "+OK",
			wantErr: ErrIncomplete,
		},
		{
			name:    "incomplete integer",
			input:   ":123",
			wantErr: ErrIncomplete,
		},
		{
			name:    "incomplete bulk string",
			input:   "$5\r\nhel",
			wantErr: ErrIncomplete,
		},
		{
			name:    "incomplete array",
			input:   "*2\r\n$3\r\nGET\r\n",
			wantErr: ErrIncomplete,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: ErrNoData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode([]byte(tt.input))

			// Check error
			if err != tt.wantErr {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			// If an error is expected, don't compare values.
			if tt.wantErr != nil {
				return
			}

			// Check decoded value.
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"expected %#v, got %#v",
					tt.expected,
					got,
				)
			}
		})
	}
}