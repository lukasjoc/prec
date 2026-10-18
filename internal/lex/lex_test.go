package lex

import (
	"crypto/sha256"
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

	var testdata []string
	if err := json.Unmarshal(data, &testdata); err != nil {
		panic(err)
	}

	basepath := filepath.Base(*fixture)
	fixtureId := strings.TrimSuffix(basepath, filepath.Ext(basepath))
	return fixtureId, testdata
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

func TestLex(t *testing.T) {
	id, data := loadFixture(t)
	for _, input := range data {
		sum := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))[:8]
		t.Run(fmt.Sprintf("TestLex-%s-%s", id, sum), func(t *testing.T) {
			l := New([]byte(input))
			result := []string{}
			for l.peekable() {
				tok, err := l.Next()
				if err != nil {
					result = append(result, err.Error())
				} else {
					result = append(result, tok.Type.String())
				}
			}
			result_json, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			snapshotAssert(t, input, string(result_json))
		})
	}
}
