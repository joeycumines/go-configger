package main

import (
	"testing"
	"runtime"
	"path"
	"os/exec"
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
}
`,
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
}
`,
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
}
`,
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
two="22"
`,
			Code: 0,
		},
		{
			Args: []string{
			},
			Expected: ``,
			Code:     CodeNoTargets,
		},
	}

	for _, testCase := range testCases {
		args := []string{`run`, path.Join(pkgPath, `main.go`)}
		cmd := exec.Command(`go`, append(args, testCase.Args...)...)

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if testCase.Code != 0 {
			// TODO: exit code
			if err == nil {
				t.Error("expected error code ", testCase.Code)
			}
			continue
		}

		if err != nil {
			t.Error(outputStr)
			continue
		}

		if testCase.Expected != outputStr {
			t.Fatalf("expected output != actual\nEXPECTED:\n%s\nACTUAL:\n%s", testCase.Expected, outputStr)
		}
	}
}
