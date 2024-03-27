package main

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/lukasjoc/prec/internal/sexpr"
	"golang.org/x/term"
)

func main() {
	old, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(fmt.Errorf("could not enter raw mode: %w", err))
	}
	defer term.Restore(int(os.Stdin.Fd()), old)

	vt := term.NewTerminal(os.Stdin, "> ")
	for {
		rawLine, err := vt.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}
		line := strings.Trim(rawLine, " ")
		if line == "help" {
			fmt.Println("Example: (+ 1 2)")
			fmt.Print("\033[2K\r")
			fmt.Println("quit, exit exit the repl")
			fmt.Print("\033[2K\r")
			fmt.Println("help print help")
			fmt.Print("\033[2K\r")
			continue
		}
		if line == "quit" || line == "exit" {
			break
		}
		builder := sexpr.NewBuilder(line)
		s, err := builder.Build()
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Print("\033[2K\r")
			continue
		}
		s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
			if s == nil {
				return
			}
			fmt.Printf("%s%s\n", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String())
			fmt.Print("\033[2K\r")
		}, &sexpr.SExprVisitCtx{Depth: 0})
		val, err := s.Eval()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Print("\033[2K\r")
		}
		if val != nil {
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
			fmt.Println(val.Text('f', prec))
			fmt.Print("\033[2K\r")
		}
	}
}
