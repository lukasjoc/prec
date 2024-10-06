package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/lukasjoc/prec/internal/sexpr"
	"golang.org/x/term"
)

var (
	verbose     = flag.Bool("v", false, "Be verbose with the output")
	evalContent = flag.String("e", "", "Directly evaluate an expression and print the result to stdout")
)

func openRepl() {
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
		if verbose != nil && *verbose == true {
			s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
				if s == nil {
					return
				}
				fmt.Printf("%s%s\n", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String())
				fmt.Print("\033[2K\r")
			}, &sexpr.SExprVisitCtx{Depth: 0})
		}
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

func main() {
	flag.Parse()
	if evalContent == nil || len(*evalContent) == 0 {
		openRepl()
	} else {
		line := *evalContent
		builder := sexpr.NewBuilder(line)
		s, err := builder.Build()
		if errors.Is(err, io.EOF) {
			os.Exit(1)
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Print("\033[2K\r")
			os.Exit(1)
		}

		if verbose != nil && *verbose == true {
			s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
				if s == nil {
					return
				}
				fmt.Printf("%s%s\n", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String())
				fmt.Print("\033[2K\r")
			}, &sexpr.SExprVisitCtx{Depth: 0})
		}
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
