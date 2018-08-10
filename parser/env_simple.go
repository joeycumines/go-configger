package parser

import (
	"io"
	"strconv"
	"fmt"
	"errors"
	"sort"
)

func EnvSimpleRead(r io.Reader) (interface{}, error) {
	result := make(map[string]interface{})
	b := make([]byte, 1024)
	c := 0
	var k, v []byte
	for {
		n, err := r.Read(b)
		for i := 0; i < n; i++ {
			if c%2 == 0 {
				if b[i] == '=' {
					c++
					continue
				}
				k = append(k, b[i])
			} else {
				if b[i] == '\n' {
					c++
					result[string(k)] = string(v)
					k = k[:0]
					v = v[:0]
					continue
				}
				v = append(v, b[i])
			}
		}
		if err == io.EOF {
			if c%2 == 1 {
				result[string(k)] = string(v)
			}
			return result, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

func EnvSimpleWrite(data interface{}, w io.Writer) error {
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
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := result[k]
		if _, err := w.Write(
			append(
				append(
					append(
						append(
							make([]byte, 0, len(k)+len(v)+2),
							k...,
						),
						"="...,
					),
					v...,
				),
				"\n"...,
			),
		); err != nil {
			return err
		}
	}
	return nil
}
