package expr

import (
	"bufio"
	"encoding/json/v2"
	"fmt"
	"os"
	"strings"
	"testing"
)

var testdata []string

func loadTestData() {
	// TODO(NOMERGE): testdata ingestion through commandline
	data, err := os.ReadFile("testdata.txt")
	if err != nil {
		panic(err)
	}

	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		testdata = append(testdata, sc.Text())
	}
}

func init() {
	loadTestData()
}

type snapshot struct {
	Input  string
	Output string
}

func snapshotAssert(t *testing.T, input string, output string) {
	a := snapshot{input, output}
	data, marshalErr := json.Marshal(a)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	t.Logf("--- SNAP:@@%s", data)
}

func TestParse(t *testing.T) {
	for index, input := range testdata {
		t.Run(fmt.Sprintf("TestParse-%d", index), func(t *testing.T) {
			p := NewParser([]byte(input))
			expr, err := p.Parse()
			if err != nil {
				snapshotAssert(t, input, err.Error())
			} else {
				expr_json, err := json.Marshal(expr)
				if err != nil {
					t.Fatal(err)
				}
				snapshotAssert(t, input, string(expr_json))
			}
		})
	}
}
