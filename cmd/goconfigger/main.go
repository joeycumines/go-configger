package main

import (
	"os"
	"gopkg.in/urfave/cli.v1"
	"github.com/joeycumines/go-configger/parser"
	"io"
	"strings"
	"fmt"
	"path"
	"bytes"
	"strconv"
)

const (
	CodeBadFormat  = 10
	CodeReadError  = 11
	CodeNoTargets  = 12
	CodeWriteError = 13
)

var (
	AppName      = `goconfigger`
	AppUsage     = `output a modified configuration file, allowing merging, modification, and conversion`
	AppUsageText = `goconfigger [OPTIONS] [--] CONFIG [CONFIG...]
    CONFIG: [--FORMAT ]PATH
      FORMAT: json|yaml|env
        if provided, FORMAT will override the file extension of PATH
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
			ok     bool
		)

		// try to parse the format via a flag? e.g. --json file_path
		if i+1 < len(args) {
			flag := []rune(strings.ToLower(args[i]))

			if len(flag) > 2 && flag[0] == '-' && flag[1] == '-' {
				format, ok = appFormats[string(flag[2:])]
			}
		}

		if ok {
			// we consumed a flag, now consume the next arg
			i++
		} else {
			// parse the format via the file path?
			ext := []rune(path.Ext(args[i]))

			if len(ext) > 0 {
				format, ok = appFormats[strings.ToLower(string(ext[1:]))]
			}

			if !ok {
				return cli.NewExitError("unable to determine the format from: "+args[i], CodeBadFormat)
			}
		}

		stats, err := os.Stat(args[i])

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("unable to read '%s': %s", args[i], err.Error()), CodeReadError)
		}

		if stats.IsDir() {
			return cli.NewExitError(fmt.Sprintf("unable to merge directory '%s'", args[i]), CodeReadError)
		}

		var r io.Reader

		r, err = os.Open(args[i])

		if err != nil {
			return cli.NewExitError(fmt.Sprintf("unable to open '%s': %s", args[i], err.Error()), CodeReadError)
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
			return cli.NewExitError(fmt.Sprintf("unable to parse file format %v': %s", input.Format, err.Error()), CodeReadError)
		}
		data = mode.Merge(data, newData)
	}

	// print the combined output
	buffer := bytes.NewBufferString("")
	if err := appParser.Write(targetFormat, data, buffer); err != nil {
		return cli.NewExitError(fmt.Sprintf("unable to output to format %v: %s", targetFormat, err.Error()), CodeWriteError)
	}
	fmt.Println(buffer.String())

	return nil
}

func appFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "format,f",
			Usage: "target format for the output, one of (json, yaml, env)",
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
		"env":  parser.Env,
		"json": parser.JSON,
		"yaml": parser.YAML,
		"yml":  parser.YAML,
	}
}

func appParser() parser.Config {
	return parser.Default
}
