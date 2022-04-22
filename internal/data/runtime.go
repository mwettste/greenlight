package data

import (
	"fmt"
	"strconv"
)

type RuntimeMin int32

func (r RuntimeMin) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)
	quotedJSONValue := strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}
