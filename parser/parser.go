package parser

import (
	"errors"
	"io"
	"fmt"
)

type Config map[Format]Def

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

func (d Def) Write(data interface{}, w io.Writer) error {
	if w == nil {
		return errors.New("nil writer")
	}
	if d.Writer == nil {
		return errors.New("undefined writer")
	}
	return d.Writer(data, w)
}

func (c Config) Read(format Format, r io.Reader) (interface{}, error) {
	def, ok := c[format]
	if !ok {
		return nil, fmt.Errorf("undefined format #%d", format)
	}
	return def.Read(r)
}

func (c Config) Write(format Format, data interface{}, w io.Writer) error {
	def, ok := c[format]
	if !ok {
		return fmt.Errorf("undefined format #%d", format)
	}
	return def.Write(data, w)
}
