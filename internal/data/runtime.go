package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type RuntimeMin int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r RuntimeMin) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}

func (r *RuntimeMin) UnmarshalJSON(data []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(data))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	splitted := strings.Split(unquotedJSONValue, " ")
	if len(splitted) != 2 || splitted[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	runtimeInMin, err := strconv.ParseInt(splitted[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	*r = RuntimeMin(runtimeInMin)
	return nil
}
