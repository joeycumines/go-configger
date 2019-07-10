package main

import (
	"bytes"
	"fmt"
	"github.com/joeycumines/go-configger/parser"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	CodeBadFormat   = 10
	CodeReadError   = 11
	CodeNoTargets   = 12
	CodeWriteError  = 13
	CodeBadArgument = 14
)

var (
	AppName      = `goconfigger`
	AppVersion   = `1.0.0`
	AppUsage     = `output a modified configuration file, allowing merging, modification, and conversion`
	AppUsageText = `goconfigger [OPTIONS] [--] CONFIG [CONFIG...]
    CONFIG: [--FORMAT [...FORMAT_ARGS]] PATH
      FORMAT: json|yaml|yml|yaml-index|yml-index|env|env-simple
        if provided, FORMAT will override the file extension of PATH
      FORMAT_ARGS:
        yaml-index|yml-index: INDEX
          allows selection of a single document from a (potentially)
          multi-document yaml file
      PATH: a valid path to a valid config file`
	AppAction  = appAction
	AppArgs    = os.Args
	AppFlags   = appFlags
	AppFormats = appFormats
	AppParser  = appParser
)

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Usage = AppUsage
	app.Flags = AppFlags()
	app.Action = AppAction
	app.UsageText = AppUsageText
	app.Version = AppVersion
	app.Run(AppArgs)
}

type mergeTarget struct {
	Format parser.Format
	Reader io.Reader
}

type Node struct {
	Path      string
	Whitelist bool
	Blacklist bool
}

type Mode map[string]*Node

func (m Mode) Included(s string) bool {
	if m == nil {
		return true
	}
	node, ok := m[s]
	if !ok {
		return true
	}
	if node.Whitelist {
		return true
	}
	return !node.Blacklist
}

func (m Mode) Define(s string) bool {
	if _, ok := m[s]; ok {
		return false
	}
	m[s] = &Node{
		Path: s,
	}
	return true
}

func (m Mode) merge(a, b interface{}, path []string) interface{} {
	switch tB := b.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		tA, _ := a.(map[string]interface{})

		if tA != nil {
			for k, vA := range tA {
				newPath := append(path, k)

				if !m.Included(strings.Join(newPath, ".")) {
					continue
				}

				result[k] = m.merge(nil, vA, newPath)
			}
		}
		if tB != nil {
			for k, vB := range tB {
				newPath := append(path, k)

				if !m.Included(strings.Join(newPath, ".")) {
					continue
				}

				vA, _ := result[k]

				result[k] = m.merge(vA, vB, newPath)
			}
		}

		return result

	case []interface{}:
		result := make([]interface{}, 0)
		tA, _ := a.([]interface{})

		if tA != nil {
			for i, vA := range tA {
				newPath := append(path, strconv.Itoa(i))

				if !m.Included(strings.Join(newPath, ".")) {
					continue
				}

				if i == len(result) {
					// append case
					result = append(result, m.merge(nil, vA, newPath))
					continue
				}

				if i >= len(result) {
					continue
				}

				result[i] = m.merge(nil, vA, newPath)
			}
		}
		if tB != nil {
			for i, vB := range tB {
				newPath := append(path, strconv.Itoa(i))

				if !m.Included(strings.Join(newPath, ".")) {
					continue
				}

				if i == len(result) {
					// append case
					result = append(result, m.merge(nil, vB, newPath))
					continue
				}

				if i >= len(result) {
					continue
				}

				result[i] = m.merge(nil, vB, newPath)
			}
		}

		return result

	default:
		return b
	}
}

func (m Mode) Merge(a, b interface{}) interface{} {
	return m.merge(a, b, make([]string, 0))
}

func appAction(c *cli.Context) error {
	appFormats := AppFormats()
	appParser := AppParser()

	inputList := make([]mergeTarget, 0)
	args := c.Args()
	var (
		targetFormat parser.Format
	)

	// handle options
	if flag := c.String("format"); flag != "" {
		if format, ok := appFormats[strings.ToLower(flag)]; ok {
			targetFormat = format
		} else {
			return cli.NewExitError("unable to determine the format from: "+flag, CodeBadFormat)
		}
	}

	// handle mode
	mode := make(Mode)
	for _, included := range c.StringSlice("whitelist") { // TODO: implement whitelist
		mode.Define(included)
		mode[included].Whitelist = true
	}
	for _, excluded := range c.StringSlice("blacklist") {
		mode.Define(excluded)
		mode[excluded].Blacklist = true
	}

	// handle args
	for i := 0; i < len(args); i++ {
		var (
			format parser.Format
			flag   string
			inc    = 1
			ok     bool
			index  int
		)

		// TODO: separator support

		// try to parse the format via a flag? e.g. --json file_path
		if i < len(args)-1 {
			if v := []rune(strings.ToLower(args[i])); len(v) > 2 && v[0] == '-' && v[1] == '-' {
				flag = string(v[2:])
				format, ok = appFormats[flag]
			}
		}
		if ok {
			// found a flag, does it require any extra args?
			switch flag {
			case `yaml-index`, `yml-index`:
				if i >= len(args)-2 {
					ok = false
				} else if v, err := strconv.Atoi(args[i+1]); err != nil || v < 0 {
					return cli.NewExitError(fmt.Sprintf("invalid %s argument index: %s", args[i], args[i+1]), CodeBadArgument)
				} else {
					inc++
					index = v
				}
			}
		}
		if !ok {
			// parse the format via the file path?
			ext := []rune(path.Ext(args[i]))
			if len(ext) > 0 {
				format, ok = appFormats[strings.ToLower(string(ext[1:]))]
			}
			if !ok {
				return cli.NewExitError("unable to determine the format from: "+args[i], CodeBadFormat)
			}
		} else {
			// successfully consumed a format, we can increment
			i += inc
		}

		stats, err := os.Stat(args[i])
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("unable to read '%s': %s", args[i], err.Error()), CodeReadError)
		}

		if stats.IsDir() {
			return cli.NewExitError(fmt.Sprintf("unable to merge directory '%s'", args[i]), CodeReadError)
		}

		var r io.Reader
		if f, err := os.Open(args[i]); err != nil {
			return cli.NewExitError(fmt.Sprintf("unable to open '%s': %s", args[i], err.Error()), CodeReadError)
		} else {
			//noinspection GoDeferInLoop
			defer f.Close()
			r = f
		}

		// read any skipped yaml items from the stream - ghetto as hell but whatever
		if index > 0 {
			var (
				d = yaml.NewDecoder(r)
				v interface{}
			)
			for x := 0; x <= index; x++ {
				if err := d.Decode(&v); err != nil {
					return cli.NewExitError(fmt.Sprintf("unable to decode yaml at %d of '%s': %s", x, args[i], err.Error()), CodeReadError)
				}
			}
			b, err := yaml.Marshal(v)
			if err != nil {
				return cli.NewExitError(fmt.Sprintf("unable to encode yaml at %d of '%s': %s", index, args[i], err.Error()), CodeReadError)
			}
			r = bytes.NewBuffer(b)
		}

		inputList = append(inputList, mergeTarget{format, r})
	}

	if len(inputList) <= 0 {
		return cli.NewExitError("at least one target config must be provided", CodeNoTargets)
	}

	if targetFormat == parser.Auto {
		targetFormat = inputList[0].Format
	}

	// merge, and apply options
	var data interface{}
	for _, input := range inputList {
		// read the file
		newData, err := appParser.Read(input.Format, input.Reader)
		if err != nil {
			return cli.NewExitError(fmt.Sprintf("unable to parse file format %v: %s", input.Format, err.Error()), CodeReadError)
		}
		data = mode.Merge(data, newData)
	}

	// print the combined output
	buffer := bytes.NewBufferString("")
	if err := appParser.Write(targetFormat, data, buffer); err != nil {
		return cli.NewExitError(fmt.Sprintf("unable to output to format %v: %s", targetFormat, err.Error()), CodeWriteError)
	}
	fmt.Print(buffer.String())

	return nil
}

func appFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "format,f",
			Usage: "target format for the output, one of (json, yaml, env, yml, env-simple)",
		},
		cli.StringSliceFlag{
			Name:  "whitelist,include,i,w",
			Usage: "whitelisted paths (dot notation) will always be included",
		},
		cli.StringSliceFlag{
			Name:  "blacklist,excluded,e,b",
			Usage: "blacklisted paths (dot notation) will be excluded unless whitelisted",
		},
	}
}

func appFormats() map[string]parser.Format {
	return map[string]parser.Format{
		"env":        parser.Env,
		"json":       parser.JSON,
		"yaml":       parser.YAML,
		"yml":        parser.YAML,
		"yaml-index": parser.YAML,
		"yml-index":  parser.YAML,
		"env-simple": parser.EnvSimple,
	}
}

func appParser() parser.Config {
	return parser.Default
}
