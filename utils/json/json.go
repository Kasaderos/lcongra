package json

import (
	"encoding/json"
	"errors"
)

var (
	ErrJSONUnmarshal = errors.New("can't unmarshal json")
	ErrNumConversion = errors.New("can't convert to number")
)

func Unmarshal(data []byte, s interface{}) error {
	return json.Unmarshal(data, s)
}
