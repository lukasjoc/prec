package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/lukasjoc/prec/internal/sexpr"
)

func printPreamble() {
	fmt.Println("prec v1")
	fmt.Println("The precision calculator with the lispy dialect.")
}

func main() {
	dumper := sexpr.NewDumper()
	stdin := bufio.NewReader(os.Stdin)
	printPreamble()
	for {
		fmt.Printf("; ")
		raw, _ := stdin.ReadString('\n')
		line := strings.TrimSpace(raw)
		if len(line) == 0 {
			continue
		}
		builder := sexpr.NewBuilder(line)
		s, err := builder.Build()
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("ERROR: could not build sexpr from `%s` %v\n", line, err)
			continue
		}
		dumper.Dump(&s)
		val, err := s.Eval()
		if err != nil {
			fmt.Printf("ERROR: could not evaluate sexpr from `%s` %v\n", line, err)
			continue
		}
		if val == nil {
			continue
		}
		prec := int(val.Prec())
		acc := val.Acc()
		if acc == big.Exact {
			fmt.Printf("%s ", acc)
			if val.IsInt() {
				prec = 0
			} else {
				prec = 1
			}
		}
		fmt.Printf("%s\n", val.Text('f', prec))
	}
}
