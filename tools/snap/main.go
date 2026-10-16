package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type InputRow struct {
	Action string
	Output string
	Test   string
}

func main() {
	stdin := bufio.NewScanner(os.Stdin)

	for stdin.Scan() {
		var row InputRow
		if err := json.Unmarshal(stdin.Bytes(), &row); err != nil {
			fmt.Fprintf(os.Stderr, "snap: cannot parse line data: error: %v\n", err)
			break
		}
		fmt.Fprintf(os.Stdout, "snap(row): %#v\n", row)

		if stdin.Err() != nil {
			break
		}
	}

}
