package parser

import (
	"errors"
	"io"
	"fmt"
)

type Config map[int]Def

type Def struct {
	Reader
	Writer
}

var Default Config

func init() {
	Default = Config{
		Auto: Def{// TODO: implement this
			Reader: nil,
			Writer: nil,
		},
		JSON: Def{
			Reader: JSONRead,
			Writer: JSONWrite,
		},
		YAML: Def{
			Reader: YAMLRead,
			Writer: YAMLWrite,
		},
		Env: Def{
			Reader: EnvRead,
			Writer: EnvWrite,
		},
	}
}

func (d Def) Read(r io.Reader) (interface{}, error) {
	if r == nil {
		return nil, errors.New("nil reader")
	}
	if d.Reader == nil {
		return nil, errors.New("undefined reader")
	}
	return d.Reader(r)
}

func (d Def) Write(data interface{}, w io.Writer) (int, error) {
	if w == nil {
		return 0, errors.New("nil writer")
	}
	if d.Writer == nil {
		return 0, errors.New("undefined writer")
	}
	return d.Write(data, w)
}

func (c Config) Read(format int, r io.Reader) (interface{}, error) {
	def, ok := c[format]
	if !ok {
		return nil, fmt.Errorf("undefined format #%d", format)
	}
	return def.Read(r)
}

func (c Config) Write(format int, data interface{}, w io.Writer) (int, error) {
	def, ok := c[format]
	if !ok {
		return 0, fmt.Errorf("undefined format #%d", format)
	}
	return def.Write(data, w)
}
