package snap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var snapshotsDir *string = nil

func SetDir(value string) {
	snapshotsDir = &value
}

type SnapshotMeta struct {
	Description string `json:"description,omitempty"`
}

type Snapshot struct {
	t        *testing.T
	Metadata SnapshotMeta `json:"metadata"`
	Output   string       `json:"output"`
}

func ParseSnapshot(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func MakeSnapshotFilename(testname string) string {
	return fmt.Sprintf("%s.snap", strings.Replace(testname, "/", "-", 1))
}

func LoadSnapshot(testname string, snapshotsDir string) (*Snapshot, error, bool) {
	path := filepath.Join(snapshotsDir, testname)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, false
		}
		return nil, err, false
	}
	s, err := ParseSnapshot(data)
	return s, err, true
}

func WriteSnapshot(testname string, snapshotsDir string, s *Snapshot) error {
	data, err := json.Marshal(s)
	path := filepath.Join(snapshotsDir, testname)
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func New(t *testing.T) *Snapshot {
	s := new(Snapshot)
	s.t = t
	return s
}

func (s *Snapshot) WithDescription(v string) *Snapshot {
	s.t.Helper()
	s.Metadata.Description = v
	return s
}

// @private allowed only within this file
func normalizeTestname(name string) string {
	return strings.Replace(name, "/", "-", 1)
}

// @private allowed only within this file
func assertSnapshot(t *testing.T, incoming *Snapshot) {
	t.Helper()

	testname := fmt.Sprintf("%s.snap", normalizeTestname(t.Name()))

	existing, err, found := LoadSnapshot(testname, *snapshotsDir)
	if err != nil {
		t.Fatal(err)
	}

	if !found {
		newtestname := fmt.Sprintf("%s.snap.added", normalizeTestname(t.Name()))
		if err := WriteSnapshot(newtestname, *snapshotsDir, incoming); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("snap: new snapshot: %s", newtestname)
	} else {
		modifiedtestname := fmt.Sprintf("%s.snap.modified", normalizeTestname(t.Name()))
		if existing.Output == incoming.Output {
			// PASSED: remove modified chunk
			os.Remove(filepath.Join(*snapshotsDir, modifiedtestname))
			return
		}
		if err := WriteSnapshot(modifiedtestname, *snapshotsDir, incoming); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("snap: modified snapshot: %s", modifiedtestname)
		// TODO(NOMERGE): delete needs to be handled as well
		// it needs to handle deleting .snap and any other snapshot state
	}
}

func (s *Snapshot) Assert(output string) {
	s.t.Helper()
	s.Output = output
	assertSnapshot(s.t, s)
}
