package expr

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// TODO(NOMERGE): dont need a json for the testdata
var testdata = []struct {
	Input string `json:"input"`
	// snapshot will take care of the expected and if an error is reported
	// Expected string `json:"expected"`
	// Error    bool   `json:"error"`
}{}

func loadTestData() {
	// TODO(NOMERGE): testdata ingestion through commandline
	data, err := os.ReadFile("testdata.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &testdata); err != nil {
		panic(err)
	}
}

func init() {
	loadTestData()
}

type snapshotAssertion struct {
	Input  string
	Output string
	Err    any
}

func snapshotAssert(t *testing.T, input string, output string, assertionErr error) {
	a := snapshotAssertion{input, output, nil}
	if assertionErr != nil {
		a.Err = assertionErr.Error()
	}
	data, marshalErr := json.Marshal(a)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	t.Logf("--- SNAP:@@%s", data)
}

func TestParse(t *testing.T) {
	for index, entry := range testdata {
		t.Run(fmt.Sprintf("TestParse-%d", index), func(t *testing.T) {
			input := []byte(entry.Input)
			p := NewParser(input)
			expr, err := p.Parse()
			if err != nil {
				snapshotAssert(t, entry.Input, "", err)
			} else {
				expr_json, err := json.Marshal(expr)
				if err != nil {
					t.Fatal(err)
				}
				snapshotAssert(t, entry.Input, string(expr_json), nil)
			}

		})
	}
}
