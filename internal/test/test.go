package test

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lukasjoc/prec/snap"
)

// TODO(NOMERGE): way to provide defaults
var fixture = flag.String("fixture", "", "test fixture to use for all test runs")
var snapshotsDir = flag.String("snapshots-dir", "", "output directory for the snapshot files")

func loadTest() ([]string, error) {
	{
		if snapshotsDir == nil || len(*snapshotsDir) == 0 {
			return nil, fmt.Errorf("loadTest: missing snapshots-dir")
		}
		fifo, err := os.Stat(*snapshotsDir)
		if err != nil {
			return nil, fmt.Errorf("loadTest: invalid snapshots-dir: %w", err)
		}
		if !fifo.IsDir() {
			return nil, fmt.Errorf("loadTest: invalid snapshots-dir: not a dir")
		}
		snap.SetDir(*snapshotsDir)
	}

	buf, err := os.ReadFile(*fixture)
	if err != nil {
		return nil, fmt.Errorf("loadTest: missing fixture: %w", err)
	}
	var testdata []string
	if err := json.Unmarshal(buf, &testdata); err != nil {
		return nil, fmt.Errorf("loadTest: invalid fixture: %w", err)
	}
	return testdata, nil
}

func makeTestname(t *testing.T, input string) string {
	basename, _, _ := strings.Cut(t.Name(), "/")
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))[:8]
	basepath := filepath.Base(*fixture)
	fixtureId := strings.TrimSuffix(basepath, filepath.Ext(basepath))
	return fmt.Sprintf("%s-%s-%s", basename, fixtureId, sum)
}
