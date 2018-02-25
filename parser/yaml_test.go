package parser

import (
	"testing"
)

func YAMLTestCases() []*RWTestCase {
	return []*RWTestCase{
		{
			Name: `int`,
			Raw: `
123

`,
			Clean: `123
`,
			Parsed: float64(123),
			Reader: YAMLRead,
			Writer: YAMLWrite,
		},
		{
			Name: `json empty`,
			Raw: `
   {

}

`,
			Clean: `{}
`,
			Parsed: map[string]interface{}{},
			Reader: YAMLRead,
			Writer: YAMLWrite,
		},
		{
			Name: `yaml complex`,
			Raw: `one: 213.23
two:
  - 2
  - true
  - "ONE"
  - "2"
five:
  six: {}
  seven:
    eight: -9123
`,
			Clean: `five:
  seven:
    eight: -9123
  six: {}
one: 213.23
two:
- 2
- true
- ONE
- "2"
`,
			Parsed: map[string]interface{}{
				"one": float64(213.23),
				"two": []interface{}{
					float64(2),
					true,
					"ONE",
					"2",
				},
				"five": map[string]interface{}{
					"six": map[string]interface{}{},
					"seven": map[string]interface{}{
						"eight": float64(-9123),
					},
				},
			},
			Reader: YAMLRead,
			Writer: YAMLWrite,
		},
	}
}

func TestYAMLRead(t *testing.T) {
	testCases := YAMLTestCases()
	for _, testCase := range testCases {
		if err := testCase.Read(); err != nil {
			t.Errorf("[%s] %v", testCase.Name, err)
		}
	}
}

func TestYAMLWrite(t *testing.T) {
	testCases := YAMLTestCases()
	for _, testCase := range testCases {
		if err := testCase.Write(); err != nil {
			t.Errorf("[%s] %v", testCase.Name, err)
		}
	}
}

func TestYAMLReadWriteRead(t *testing.T) {
	testCases := YAMLTestCases()
	for _, testCase := range testCases {
		if err := testCase.Read(); err != nil {
			t.Error("READ failure: ", err)
		}
		if err := testCase.Write(); err != nil {
			t.Error("WRITE failure: ", err)
		}
		if err := testCase.Read(); err != nil {
			t.Error("READ failure: ", err)
		}
	}
}
