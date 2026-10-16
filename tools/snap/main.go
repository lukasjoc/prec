package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type InputRow struct {
	Action string
	Output string
	Test   string
}

type snapshot struct {
	Input  string
	Output string
	Err    any
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

func main() {
	// os.RemoveAll(".snapdata")
	os.Mkdir(".snapdata", 0755)

	stdin := bufio.NewScanner(os.Stdin)

	tty, err := os.Open("/dev/tty")
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	confirm := bufio.NewReader(tty)

	for stdin.Scan() {
		var row InputRow
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
					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					if incoming.Err != nil {
						fmt.Fprintf(os.Stdout, "Output: %v\n", incoming.Err)
					} else {
						// TODO(NOMERGE): pretty printed, indented
						fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)
					}

					fmt.Fprintf(os.Stdout, "Accept new snapshot [y,n]? ")
					ans, err := confirm.ReadString('\n')
					if err != nil {
						fmt.Fprintf(os.Stderr, "snap: cannot read tty: error: %v\n", err)
						break
					}
					if ans == "y\n" || ans == "n\n" {
						if ans == "y\n" {
							data, err := json.Marshal(incoming)
							name := fmt.Sprintf(".snapdata/%s.snap", strings.Replace(row.Test, "/", "-", 1))
							err = os.WriteFile(name, data, 0644)
							if err != nil {
								fmt.Fprintf(os.Stderr, "snap: cannot write snapshot: error: %v\n", err)
								break
							}
						} else {
							continue
						}
					} else {
						panic("handle wrong input")
					}
				} else {
					if existing.Input == incoming.Input && existing.Err != nil && incoming.Err != nil && existing.Err == incoming.Err {
						continue
					}
					if existing.Input == incoming.Input && existing.Output == incoming.Output {
						continue
					}

					fmt.Fprintf(os.Stdout, "Name: %s\n", row.Test)

					fmt.Fprintf(os.Stdout, "EXISTING ==>\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", existing.Input)
					if existing.Err != nil {
						fmt.Fprintf(os.Stdout, "Output: %v\n", existing.Err)
					} else {
						// TODO(NOMERGE): pretty printed, indented
						fmt.Fprintf(os.Stdout, "Output: %s\n", existing.Output)
					}

					fmt.Fprintf(os.Stdout, "<== INCOMING\n")
					fmt.Fprintf(os.Stdout, "Input: %s\n", incoming.Input)
					if incoming.Err != nil {
						fmt.Fprintf(os.Stdout, "Output: %v\n", incoming.Err)
					} else {
						// TODO(NOMERGE): pretty printed, indented
						fmt.Fprintf(os.Stdout, "Output: %s\n", incoming.Output)
					}

					fmt.Fprintf(os.Stdout, "Accept updates [y,n]? ")
					ans, err := confirm.ReadString('\n')
					if err != nil {
						fmt.Fprintf(os.Stderr, "snap: cannot read tty: error: %v\n", err)
						break
					}
					if ans == "y\n" || ans == "n\n" {
						if ans == "y\n" {
							panic("handle accepted updates")
							// data, err := json.Marshal(incoming)
							// name := fmt.Sprintf(".snapdata/%s.snap", strings.Replace(row.Test, "/", "-", 1))
							// err = os.WriteFile(name, data, 0644)
							// if err != nil {
							// 	fmt.Fprintf(os.Stderr, "snap: cannot write snapshot: error: %v\n", err)
							// 	break
							// }
						} else {
							continue
						}
					} else {
						panic("handle wrong input")
					}
				}
				fmt.Fprintf(os.Stdout, "\n")
			}
		}

		if stdin.Err() != nil {
			break
		}
	}

}
