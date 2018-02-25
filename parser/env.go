package parser

import (
	"io"
	"github.com/joho/godotenv"
	"errors"
	"strconv"
	"fmt"
)

func EnvRead(r io.Reader) (interface{}, error) {
	m, err := godotenv.Parse(r)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	if m != nil {
		for k, v := range m {
			result[k] = v
		}
	}
	return result, nil
}

func EnvWrite(data interface{}, w io.Writer) error {
	m, ok := data.(map[string]interface{})
	if !ok {
		return errors.New(".env only supports maps")
	}
	result := make(map[string]string)
	if m != nil {
		for k, v := range m {
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case bool:
				if t {
					result[k] = "true"
				} else {
					result[k] = "false"
				}
			case float64:
				result[k] = strconv.FormatFloat(t, 'f', -1, 64)
			case string:
				result[k] = t
			default:
				return fmt.Errorf("unsupported type %T for property '%s'", v, k)
			}
		}
	}
	str, err := godotenv.Marshal(result)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(str))
	return err
}
