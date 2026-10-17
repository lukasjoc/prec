package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

func loadSnapshot(testname string) (*snapshot, error, bool) {
	name := fmt.Sprintf(".snapdata/%s.snap", strings.Replace(testname, "/", "-", 1))
	data, err := os.ReadFile(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, false
		}
		return nil, err, false
	}
	s, err := parseSnapshot(data)
	return s, err, true
}

func writeSnapshot(testname string, s *snapshot) error {
	data, err := json.Marshal(s)
	name := fmt.Sprintf(".snapdata/%s.snap", strings.Replace(testname, "/", "-", 1))
	err = os.WriteFile(name, data, 0644)
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
		case "":
			return true, nil
		case "y":
			return true, nil
		case "n":
			return true, nil
		}
		fmt.Fprintf(os.Stdout, "error: invalid reply: \"%s\"\n", ans)
	}
}

func main() {
	// os.RemoveAll(".snapdata")
	os.Mkdir(".snapdata", 0755)

	stdin := bufio.NewScanner(os.Stdin)

	tty, err := os.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	ttyin := bufio.NewReader(tty)

	for stdin.Scan() {
		var row struct {
			Action string
			Output string
			Test   string
		}

		if err := json.Unmarshal(stdin.Bytes(), &row); err != nil {
			fmt.Fprintf(os.Stderr, "snap: cannot parse line data: error: %v\n", err)
			break
		}

		if row.Action == "output" {
			// parse_test.go:48: --- SNAP:
			output := strings.Trim(row.Output, " \n\r")
			if strings.Contains(output, "--- SNAP:@@") {
				parts := strings.Split(output, "@@")
				incoming, err := parseSnapshot([]byte(parts[len(parts)-1]))
				if err != nil {
					fmt.Fprintf(os.Stderr, "snap: cannot parse snapshot: error: %v\n", err)
					break
				}

				existing, err, found := loadSnapshot(row.Test)
				if err != nil {
					fmt.Fprintf(os.Stderr, "snap: cannot load snapshot: error: %v\n", err)
					break
				}
				if !found {
					// TODO(NOMERGE): pretty printed, indented
					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)

					ans, err := ttyPrompt(ttyin, "Accept new snapshot?")
					if err != nil {
						fmt.Fprintf(os.Stderr, "snap: cannot read from tty: error: %v\n", err)
						break
					}
					if ans {
						if err := writeSnapshot(row.Test, incoming); err != nil {
							fmt.Fprintf(os.Stderr, "snap: cannot write snapshot: error: %v\n", err)
							break
						}
					}
				} else {
					if existing.Input == incoming.Input && existing.Output == incoming.Output {
						continue
					}
					// TODO(NOMERGE): pretty printed, indented
					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)
					fmt.Fprintf(os.Stdout, "EXISTING ==>\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", existing.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", existing.Output)
					fmt.Fprintf(os.Stdout, "<== INCOMING\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)

					ans, err := ttyPrompt(ttyin, "Accept updates?")
					if err != nil {
						fmt.Fprintf(os.Stderr, "snap: cannot read from tty: error: %v\n", err)
						break
					}
					if ans {
						if err := writeSnapshot(row.Test, incoming); err != nil {
							fmt.Fprintf(os.Stderr, "snap: cannot write snapshot: error: %v\n", err)
							break
						}
					}
				}
			}
		}

		if stdin.Err() != nil {
			break
		}
	}

}
