package schemaregistry

import (
	"strconv"
	"strings"
)

// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#errors
type Error struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"error_code"`
	Message    string `json:"message"`
}

func newError(statusCode int) *Error {
	return &Error{
		StatusCode: statusCode,
	}
}

func (e Error) String() string {

	b := strings.Builder{}
	b.WriteString(strconv.Itoa(e.StatusCode))
	b.WriteByte(':')
	b.WriteString(strconv.Itoa(e.Code))
	b.WriteByte(' ')
	b.WriteString(e.Message)

	return b.String()
}

func (e Error) Error() string {
	return e.String()
}
