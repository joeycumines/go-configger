package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-test/deep"
	"io/ioutil"
	"path"
	"runtime"
	"strings"
)

const testDataDir = "testdata"

var (
	pkgPath string
)

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to load pkgPath")
	}
	pkgPath = path.Dir(file)
}

func readFile(file string) string {
	b, err := ioutil.ReadFile(path.Join(pkgPath, testDataDir, file))
	if err != nil {
		panic(err)
	}
	return string(b)
}

type RWTestCase struct {
	Name   string
	Raw    string
	Clean  string
	Parsed interface{}
	Reader Reader
	Writer Writer
}

func (c *RWTestCase) Read() error {
	o, err := c.Reader(bytes.NewBufferString(c.Raw))
	if err != nil {
		return err
	}
	if diff := deep.Equal(c.Parsed, o); diff != nil {
		name := c.Name
		if name == "" {
			name = "read test"
		}
		return fmt.Errorf("%s failed, diff: %s", name, strings.Join(diff, ", "))
	}
	if v, err := copyViaJSON(o); err != nil {
		return err
	} else if diff := deep.Equal(c.Parsed, v); diff != nil {
		name := c.Name
		if name == "" {
			name = "read test"
		}
		return fmt.Errorf("%s failed, not json compatible, diff: %s", name, strings.Join(diff, ", "))
	}
	c.Parsed = o
	return nil
}

func (c *RWTestCase) Write() error {
	buffer := bytes.NewBufferString("")
	if err := c.Writer(c.Parsed, buffer); err != nil {
		return err
	}
	actual := buffer.String()
	if actual != c.Clean {
		aSize := len([]rune(actual))
		cSize := len([]rune(c.Clean))
		notes := ""
		if aSize != cSize {
			notes += fmt.Sprintf(", actual len(%d) != expected len(%d)", aSize, cSize)
		}
		name := c.Name
		if name == "" {
			name = "read test"
		}
		return fmt.Errorf("%s failed, expected written != actual written%s\nEXPECTED:\n%s\nACTUAL:\n%s", name, notes, c.Clean, actual)
	}
	c.Raw = actual
	return nil
}

func parseJSON(s string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(s), result); err != nil {
		panic(err)
	}
	return result
}

func copyViaJSON(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	v = nil
	err = json.Unmarshal(b, &v)
	return v, err
}
