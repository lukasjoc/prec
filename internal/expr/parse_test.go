package expr

import (
	"bufio"
	"encoding/json/v2"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var fixture = flag.String("fixture", "", "test fixture to use for all test runs")

func loadFixture(t *testing.T) (string, []string) {
	t.Helper()
	data, err := os.ReadFile(*fixture)
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Split(bufio.ScanLines)
	lines := []string{}
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	basepath := filepath.Base(*fixture)
	fixtureId := strings.TrimSuffix(basepath, filepath.Ext(basepath))
	return fixtureId, lines
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
	id, data := loadFixture(t)
	for index, input := range data {
		t.Run(fmt.Sprintf("TestParse-%s-%d", id, index), func(t *testing.T) {
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
