package parser

import (
	"testing"
)

func JSONTestCases() []*RWTestCase {
	return []*RWTestCase{
		{
			Raw: `
  { 


  } 

`,
			Clean:  `{}`,
			Parsed: map[string]interface{}{},
			Reader: JSONRead,
			Writer: JSONWrite,
		},
		{
			Raw: `
   null

`,
			Clean:  `null`,
			Parsed: nil,
			Reader: JSONRead,
			Writer: JSONWrite,
		},
		{
			Raw: `{ "one": 1 }`,
			Clean: `{
  "one": 1
}`,
			Parsed: map[string]interface{}{
				"one": float64(1),
			},
			Reader: JSONRead,
			Writer: JSONWrite,
		},
	}
}

func TestJSONRead(t *testing.T) {
	testCases := JSONTestCases()
	for _, testCase := range testCases {
		if err := testCase.Read(); err != nil {
			t.Error(err)
		}
	}
}

func TestJSONWrite(t *testing.T) {
	testCases := JSONTestCases()
	for _, testCase := range testCases {
		if err := testCase.Write(); err != nil {
			t.Error(err)
		}
	}
}

func TestJSONReadWriteRead(t *testing.T) {
	testCases := JSONTestCases()
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
