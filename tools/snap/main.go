package main

import (
	"bufio"
	"encoding/json/v2"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type snapshot struct {
	Input  string
	Output string
}

func parseSnapshot(data []byte) (*snapshot, error) {
	var s snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func makeSnapshotFilename(testname string) string {
	return fmt.Sprintf("%s.snap", strings.Replace(testname, "/", "-", 1))
}

func loadSnapshot(testname string, snapshotsDir string) (*snapshot, error, bool) {
	path := filepath.Join(snapshotsDir, makeSnapshotFilename(testname))
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, false
		}
		return nil, err, false
	}
	s, err := parseSnapshot(data)
	return s, err, true
}

func writeSnapshot(testname string, snapshotsDir string, s *snapshot) error {
	data, err := json.Marshal(s)
	path := filepath.Join(snapshotsDir, makeSnapshotFilename(testname))
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ttyPrompt(rd *bufio.Reader, prompttext string) (bool, error) {
	for {
		fmt.Fprintf(os.Stdout, "%s [Yn]? ", prompttext)
		anstext, err := rd.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("ttyPrompt: could not read from tty: error: %w\n", err)
		}
		ans := strings.Trim(anstext, " \n\r")
		switch ans {
		case "", "y", "Y":
			return true, nil
		case "n", "N":
			return false, nil
		}
		fmt.Fprintf(os.Stdout, "error: invalid reply: \"%s\"\n", ans)
	}
}

func bailWithUsage(err error) {
	program := os.Args[0]
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nhelp: %s --help\n", program)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "usage: %s [--snapshots-dir,--help]\n", program)
	fmt.Fprintf(os.Stderr, "    --snapshots-dir   output directory for the snapshot files (default: .snapdata)\n")
	fmt.Fprintf(os.Stderr, "    -i, --interactive show prompts for changes\n")
	fmt.Fprintf(os.Stderr, "    -h, --help        show this help\n")
	fmt.Fprintf(os.Stderr, "example: go test -json ./... | %s\n", program)
	fmt.Fprintf(os.Stderr, "example: go test -json ./... | %s -i\n", program)
	os.Exit(0)
}

func main() {
	flags := flag.NewFlagSet("options", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	snapshotsDir := flags.String("snapshots-dir", ".snapdata", "")

	var modeInteractive bool
	flags.BoolVar(&modeInteractive, "i", false, "")
	flags.BoolVar(&modeInteractive, "interactive", false, "")

	if err := flags.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			bailWithUsage(nil)
		}
		bailWithUsage(err)
	}

	if modeInteractive && snapshotsDir != nil {
		if err := os.Mkdir(*snapshotsDir, 0755); err != nil && !os.IsExist(err) {
			fmt.Fprintf(os.Stderr, "snap: could not create snapshots dir: error: %v", err)
			os.Exit(1)
		}
	}

	var ttyin *bufio.Reader
	if modeInteractive {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintf(os.Stderr, "snap: could not get tty: error: %v", err)
			os.Exit(1)
		}
		defer tty.Close()

		ttyin = bufio.NewReader(tty)
	}

	failed := false

	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		var row struct {
			Action string
			Output string
			Test   string
		}

		if err := json.Unmarshal(stdin.Bytes(), &row); err != nil {
			fmt.Fprintf(os.Stderr, "snap: could not parse line data: error: %v\n", err)
			break
		}

		if row.Action == "output" {
			// parse_test.go:48: --- SNAP:
			output := strings.Trim(row.Output, " \n\r")
			if strings.Contains(output, "--- SNAP:@@") {
				parts := strings.Split(output, "@@")
				incoming, err := parseSnapshot([]byte(parts[len(parts)-1]))
				if err != nil {
					fmt.Fprintf(os.Stderr, "snap: could not parse snapshot: error: %v\n", err)
					break
				}

				existing, err, found := loadSnapshot(row.Test, *snapshotsDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "snap: could not load snapshot: error: %v\n", err)
					break
				}

				if found && existing.Input == incoming.Input && existing.Output == incoming.Output {
					continue
				}

				var prompttext string
				if !found {
					prompttext = "Accept new snapshot?"
					// TODO(NOMERGE): pretty printed, indented
					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)
				} else {
					prompttext = "Accept updates?"
					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)
					fmt.Fprintf(os.Stdout, "EXISTING ==>\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", existing.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", existing.Output)
					fmt.Fprintf(os.Stdout, "<== INCOMING\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)
				}

				if modeInteractive {
					ans, err := ttyPrompt(ttyin, prompttext)
					if err != nil {
						fmt.Fprintf(os.Stderr, "snap: could not read from tty: error: %v\n", err)
						break
					}
					if ans {
						if err := writeSnapshot(row.Test, *snapshotsDir, incoming); err != nil {
							fmt.Fprintf(os.Stderr, "snap: could not write snapshot: error: %v\n", err)
							break
						}
					} else {
						failed = true
					}
				} else {
					failed = true
				}
			}
		}

		if stdin.Err() != nil {
			break
		}
	}

	if failed {
		os.Exit(1)
	}
}
