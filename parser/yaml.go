package parser

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
)

func YAMLRead(r io.Reader) (interface{}, error) {
	decoder := yaml.NewDecoder(r)
	var result interface{}
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return fixYAMLToJSON(result)
}

func YAMLWrite(data interface{}, w io.Writer) error {
	result, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	if _, err := w.Write(result); err != nil {
		return err
	}
	return nil
}

func fixYAMLToJSON(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return t, nil
	case int:
		return float64(t), nil
	case int8:
		return float64(t), nil
	case int16:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case uint:
		return float64(t), nil
	case uint8:
		return float64(t), nil
	case uint16:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return float64(t), nil
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		if t != nil {
			for k, v := range t {
				if next, err := fixYAMLToJSON(v); err != nil {
					return nil, err
				} else {
					m[fmt.Sprintf("%v", k)] = next
				}
			}
		}
		return m, nil
	case []interface{}:
		s := make([]interface{}, len(t))
		for i, v := range t {
			if next, err := fixYAMLToJSON(v); err != nil {
				return nil, err
			} else {
				s[i] = next
			}
		}
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
