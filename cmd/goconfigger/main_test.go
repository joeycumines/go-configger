package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

type testCase struct {
	Args     []string
	Expected string
	Code     int
}

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

func TestSuccess(t *testing.T) {
	testCases := []testCase{
		{
			Args: []string{
				pkgPath + `/testdata/example.json`,
				pkgPath + `/testdata/example.yaml`,
			},
			Expected: `{
  "array": [
    11,
    22,
    3
  ],
  "nested": {
    "another": {},
    "more": [
      0.1,
      0.2
    ],
    "overridden": "something"
  },
  "unique": true,
  "unique_yaml": 14.64
}`,
			Code: 0,
		},
		{
			Args: []string{
				pkgPath + `/testdata/example.yaml`,
				pkgPath + `/testdata/example.json`,
			},
			Expected: `array:
- 1
- 2
- 3
nested:
  another: {}
  more:
  - 0.1
  - 0.2
  overridden: 9.5
unique: true
unique_yaml: 14.64
`,
			Code: 0,
		},
		{
			Args: []string{
				`--format`,
				`JSON`,
				pkgPath + `/testdata/example.yaml`,
				pkgPath + `/testdata/example.json`,
				pkgPath + `/testdata/example.yaml`,
			},
			Expected: `{
  "array": [
    11,
    22,
    3
  ],
  "nested": {
    "another": {},
    "more": [
      0.1,
      0.2
    ],
    "overridden": "something"
  },
  "unique": true,
  "unique_yaml": 14.64
}`,
			Code: 0,
		},
		{
			Args: []string{
				`-i`,
				`nested.more.0`,
				`-e`,
				`nested`,
				`-e`,
				`unique_yaml`,
				`-e`,
				`array.2`,
				pkgPath + `/testdata/example.json`,
				pkgPath + `/testdata/example.yaml`,
			},
			Expected: `{
  "array": [
    11,
    22
  ],
  "unique": true
}`,
			Code: 0,
		},
		{
			Args: []string{
				pkgPath + `/testdata/simple.env`,
				pkgPath + `/testdata/simple.json`,
				pkgPath + `/testdata/simple.yml`,
			},
			Expected: `four="34"
one="11"
three="33"
two="22"`,
			Code: 0,
		},
		{
			Args: []string{
				pkgPath + `/testdata/simple-noend.env-simple`,
				pkgPath + `/testdata/simple.json`,
				pkgPath + `/testdata/simple.yml`,
			},
			Expected: `four=34
one=11
three=33
two=22
`,
			Code: 0,
		},
		{
			Args: []string{
			},
			Expected: ``,
			Code:     CodeNoTargets,
		},
		{
			Args: []string{
				`--`,
				`--yaml-index`,
				`0`,
				pkgPath + `/testdata/example.yaml`,
				pkgPath + `/testdata/example.json`,
			},
			Expected: `array:
- 1
- 2
- 3
nested:
  another: {}
  more:
  - 0.1
  - 0.2
  overridden: 9.5
unique: true
unique_yaml: 14.64
`,
			Code: 0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yamL-index`,
				`0`,
			},
			Expected: ``,
			Code:     CodeBadFormat,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--ymL-index`,
				`0`,
			},
			Expected: ``,
			Code:     CodeBadFormat,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`-1`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: ``,
			Code:     CodeBadArgument,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				``,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: ``,
			Code:     CodeBadArgument,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`0`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `0`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yml-index`,
				`1`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `1`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yamL-index`,
				`2`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `"2"`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--ymL-index`,
				`3`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `"Three!"`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`4`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `[
  4
]`,
			Code: 0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`5`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `{
  "five": null
}`,
			Code: 0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`6`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `null`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`7`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `true`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`8`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `null`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`9`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: `null`,
			Code:     0,
		},
		{
			Args: []string{
				`-f`,
				`json`,
				`--`,
				`--yaml-index`,
				`10`,
				pkgPath + `/testdata/multi.yml`,
			},
			Expected: ``,
			Code:     CodeReadError,
		},
	}

	var bin string
	func() {
		dir, err := ioutil.TempDir(``, ``)
		if err != nil {
			t.Fatal(err)
		}
		bin = filepath.Join(dir, `goconfigger.exe`)
		cmd := exec.Command(`go`, `build`, `-v`, `-o`, bin, pkgPath)
		cmd.Dir = pkgPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}()
	defer os.Remove(bin)

	for _, testCase := range testCases {
		cmd := exec.Command(bin, testCase.Args...)

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if testCase.Code != 0 {
			if err == nil {
				t.Error("expected error code ", testCase.Code)
			} else if err, ok := err.(*exec.ExitError); ok {
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					if code := status.ExitStatus(); code != testCase.Code {
						t.Errorf("expected error code %v got %v", testCase.Code, code)
					}
				}
			}
			continue
		}

		if err != nil {
			t.Error(err)
			t.Log(outputStr)
			continue
		}

		if testCase.Expected != outputStr {
			t.Fatalf("expected output != actual\nEXPECTED:\n%s\nACTUAL:\n%s", testCase.Expected, outputStr)
		}
	}
}
