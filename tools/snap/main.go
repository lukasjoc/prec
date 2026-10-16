package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type InputRow struct {
	Action string
	Output string
	Test   string
}

type snapshotAssertion struct {
	Input      string
	Output     string
	OutputData any
	Err        any
}

func main() {
	stdin := bufio.NewScanner(os.Stdin)

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
				testdata := []byte(parts[len(parts)-1])
				// var data snapshotAssertion
				// err := json.Unmarshal(testdata, &data)
				// if err != nil {
				// 	fmt.Fprintf(os.Stderr, "snap: cannot parse output: error: %v\n", err)
				// 	break
				// }

				// if len(data.Output) > 0 {
				// 	if err := json.Unmarshal([]byte(data.Output), &data.OutputData); err != nil {
				// 		fmt.Fprintf(os.Stderr, "snap: cannot parse outputdata: error: %v\n", err)
				// 		break
				// 	}
				// }
				err := os.WriteFile(fmt.Sprintf(".snapdata/%s", strings.Replace(row.Test, "/", "-", 1)), []byte(testdata), 0600)
				if err != nil {
					fmt.Fprintf(os.Stderr, "snap: cannot write snapshot: error: %v\n", err)
					break
				}
				fmt.Fprintf(os.Stdout, "snap(row): %v\n", testdata)
			}
		}

		if stdin.Err() != nil {
			break
		}
	}

}
