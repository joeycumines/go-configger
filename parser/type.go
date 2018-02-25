package parser

import "io"

// constants for the type of file
const (
	Auto = iota
	JSON
	YAML
	Env
)

// Reader loads a given file format into memory, into the same possible types as json.Unmarshal(data, anInterface).
// Supported types:
//
//bool, for JSON booleans
//float64, for JSON numbers
//string, for JSON strings
//[]interface{}, for JSON arrays
//map[string]interface{}, for JSON objects
//nil for JSON null
type Reader func(r io.Reader) (interface{}, error)

// Writer writes out data (which is like the interface in `json.Unmarshal(data, anInterface)`), into a given format,
// via a Writer.
// Supported types:
//
//bool, for JSON booleans
//float64, for JSON numbers
//string, for JSON strings
//[]interface{}, for JSON arrays
//map[string]interface{}, for JSON objects
//nil for JSON null
type Writer func(data interface{}, w io.Writer) error
