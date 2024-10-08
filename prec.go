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

const replEndline = "\n\033[2K\r"

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
			fmt.Printf("Example: (+ 1 2)%s", replEndline)
			fmt.Printf("quit|exit exit the repl%s", replEndline)
			fmt.Printf("verbose   toggle verbose mode (enabled: %v)%s", verbose != nil && *verbose, replEndline)
			fmt.Printf("help      print help%s", replEndline)
			continue
		}
		if line == "quit" || line == "exit" {
			break
		}
		if line == "verbose" {
			if verbose != nil && *verbose {
				*verbose = false
				fmt.Printf("INFO: Verbose mode turned off%s", replEndline)
			} else {
				*verbose = true
				fmt.Printf("INFO: Verbose mode turned on%s", replEndline)
			}
			continue
		}
		builder := sexpr.NewBuilder(line)
		s, err := builder.Build()
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			fmt.Printf("Error: %v%s", err, replEndline)
			continue
		}
		if verbose != nil && *verbose {
			s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
				if s == nil {
					return
				}
				fmt.Printf("%s%s%s", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String(), replEndline)
			}, &sexpr.SExprVisitCtx{Depth: 0})
		}
		val, err := s.Eval()
		if err != nil {
			fmt.Printf("Error: %v%s", err, replEndline)
			continue
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
			fmt.Printf("%s%s", val.Text('f', prec), replEndline)
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
			os.Exit(1)
		}
		if verbose != nil && *verbose {
			s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
				if s == nil {
					return
				}
				fmt.Printf("%s%s\n", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String())
			}, &sexpr.SExprVisitCtx{Depth: 0})
		}
		val, err := s.Eval()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
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
			fmt.Printf("%s\n", val.Text('f', prec))
		}
	}
}
