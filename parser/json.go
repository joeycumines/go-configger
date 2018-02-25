package parser

import (
	"io"
	"encoding/json"
)

func JSONRead(r io.Reader) (interface{}, error) {
	decoder := json.NewDecoder(r)
	var result interface{}
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func JSONWrite(data interface{}, w io.Writer) error {
	result, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(result)
	return err
}