package parser

import (
	"testing"
)

func EnvTestCases() []*RWTestCase {
	return []*RWTestCase{
		{
			Name: `example.env`,
			Raw:  readFile("example.env"),
			Clean: `ONE="one"
QUOTED="    A  B  C  "
SPACED="A  B  C"
TWO=2
Three="four, five"`,
			Parsed: map[string]interface{}{
				"ONE":    "one",
				"TWO":    "2",
				"Three":  "four, five",
				"SPACED": "A  B  C",
				"QUOTED": "    A  B  C  ",
			},
			Reader: EnvRead,
			Writer: EnvWrite,
		},
		{
			Name:   `empty`,
			Raw:    ``,
			Clean:  ``,
			Parsed: map[string]interface{}{},
			Reader: EnvRead,
			Writer: EnvWrite,
		},
	}
}

func TestEnvRead(t *testing.T) {
	testCases := EnvTestCases()
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if err := testCase.Read(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEnvWrite(t *testing.T) {
	testCases := EnvTestCases()
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if err := testCase.Write(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEnvReadWriteRead(t *testing.T) {
	testCases := EnvTestCases()
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			if err := testCase.Read(); err != nil {
				t.Error("READ failure: ", err)
			}
			if err := testCase.Write(); err != nil {
				t.Error("WRITE failure: ", err)
			}
			if err := testCase.Read(); err != nil {
				t.Error("READ failure: ", err)
			}
		})
	}
}
