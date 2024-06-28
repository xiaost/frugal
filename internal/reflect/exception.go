package reflect

import (
	"fmt"
)

var (
	errDepthLimitExceeded = &tProtocolException{t: thrift_DEPTH_LIMIT, m: "depth limit exceeded"}
	errInvalidData        = &tProtocolException{t: thrift_INVALID_DATA, m: "invalid data"}
)

// tProtocolException implements TProtocolException of apache thrift
type tProtocolException struct {
	t int
	m string
}

// consts from in github.com/apache/thrift@v0.13.0/lib/go/thrift
const (
	thrift_UNKNOWN_PROTOCOL_EXCEPTION = 0
	thrift_INVALID_DATA               = 1
	thrift_NEGATIVE_SIZE              = 2
	thrift_SIZE_LIMIT                 = 3
	thrift_BAD_VERSION                = 4
	thrift_NOT_IMPLEMENTED            = 5
	thrift_DEPTH_LIMIT                = 6
)

// TypeId implements apache thrift TProtocolException
func (t *tProtocolException) TypeId() int { return thrift_INVALID_DATA }

// TypeID implements kitex TypeID interface
func (t *tProtocolException) TypeID() int { return thrift_INVALID_DATA }

func (e *tProtocolException) String() string { return e.m }
func (e *tProtocolException) Error() string  { return e.m }

func newRequiredFieldNotSetException(name string) error {
	return &tProtocolException{
		t: thrift_INVALID_DATA,
		m: fmt.Sprintf("required field %q is not set", name),
	}
}

func newUnknownDataTypeException(t ttype) error {
	return &tProtocolException{
		t: thrift_INVALID_DATA,
		m: fmt.Sprintf("unknown data type %d", t),
	}
}
