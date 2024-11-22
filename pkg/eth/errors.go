package eth

import (
	"encoding/json"
	"fmt"
)

// unknown service
// return fmt.Sprintf("The method %s%s%s does not exist/is not available", e.service, serviceMethodSeparator, e.method)
var MethodNotFoundErrorCode = -32601

// invalid request
var InvalidRequestErrorCode = -32600
var InvalidMessageErrorCode = -32700
var InvalidParamsErrorCode = -32602

// logic error
var CallbackErrorCode = -32000

// shutdown error
// "server is shutting down"
var ShutdownErrorCode = -32000
var ShutdownError = NewJSONRPCError(ShutdownErrorCode, "server is shutting down", nil)

func NewMethodNotFoundError(method string) *JSONRPCError {
	return NewJSONRPCError(
		MethodNotFoundErrorCode,
		fmt.Sprintf("The method %s does not exist/is not available", method),
		nil,
	)
}

func NewInvalidRequestError(message string) *JSONRPCError {
	return NewJSONRPCError(InvalidRequestErrorCode, message, nil)
}

func NewInvalidMessageError(message string) *JSONRPCError {
	return NewJSONRPCError(InvalidMessageErrorCode, message, nil)
}

func NewInvalidParamsError(message string) *JSONRPCError {
	return NewJSONRPCError(InvalidParamsErrorCode, message, nil)
}

func NewCallbackError(message string) *JSONRPCError {
	return NewJSONRPCError(CallbackErrorCode, message, nil)
}

type JSONRPCError struct {
	code    int    `json:"code"`
	message string `json:"message,omitempty"`
	err     error  `json:"details,omitempty"`
}

func NewJSONRPCError(code int, message string, err error) *JSONRPCError {
	return &JSONRPCError{
		code:    code,
		message: message,
		err:     err,
	}
}

func (err *JSONRPCError) Code() int {
	return err.code
}

func (err *JSONRPCError) Message() string {
	return err.message
}

func (err *JSONRPCError) Error() error {
	return err.err
}

// MarshalJSON implements the json.Marshaler interface.
func (d *JSONRPCError) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}
	if d.message == "" {
		return []byte("null"), nil
	}
	return json.Marshal(struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    d.code,
		Message: d.message,
	})
}

func (r *JSONRPCError) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if string(data) == "{}" {
		return nil
	}
	type ErrorData struct {
		Code    int    `json:"code"`
		Message string `json:"message",omitempty`
	}
	var resp ErrorData
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	*r = *NewJSONRPCError(
		resp.Code,
		resp.Message,
		nil,
	)

	return nil
}
