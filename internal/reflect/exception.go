package reflect

import (
	"fmt"
)

// RequiredFieldNotSetException represents a TProtocolException
// it implements TProtocolException
type requiredFieldNotSetException struct {
	message string
}

func newRequiredFieldNotSetException(name string) error {
	return &requiredFieldNotSetException{message: fmt.Sprintf("required field %q is not set", name)}
}

func (e *requiredFieldNotSetException) String() string {
	return e.message
}

func (e *requiredFieldNotSetException) Error() string {
	return e.message
}

// INVALID_DATA in github.com/apache/thrift@v0.13.0/lib/go/thrift/protocol_exception.go
const thrift_INVALID_DATA = 1

// TypeId implements apache thrift TProtocolException
func (requiredFieldNotSetException) TypeId() int { return thrift_INVALID_DATA }

// TypeID implements kitex TypeID interface
func (requiredFieldNotSetException) TypeID() int { return thrift_INVALID_DATA }
