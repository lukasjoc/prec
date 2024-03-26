package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lukasjoc/prec/internal/readline"
	"github.com/lukasjoc/prec/internal/sexpr"
)

func main() {
	old, _ := readline.GetTerminalMode(os.Stdin)
	readline.SetTerminalRawMode(os.Stdin, old)
	defer readline.ResetTerminalRawMode(os.Stdin, old)

	rl := readline.New(os.Stdin)
outer:
	for {
		input := rl.Poll()
		switch input.Key() {
		case readline.KeyArrowLeft:
			rl.MoveLeft()
		case readline.KeyArrowRight:
			rl.MoveRight()
		case readline.KeyRune:
			input := input.(readline.RuneInput)
			switch input.Value() {
			case '\r', '\n':
				line := rl.Text()
				if line == "quit" {
					rl.ResetLine()
					break outer
				}
				rl.MoveToNextLine()
				rl.ResetLine()

				builder := sexpr.NewBuilder(line)
				s, err := builder.Build()
				if errors.Is(err, io.EOF) {
					continue
				}
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					rl.ResetLine()
					continue
				}
				s.Visit(func(s *sexpr.SExpr, ctx *sexpr.SExprVisitCtx) {
					if s == nil {
						return
					}
					fmt.Printf("%s%s\n", strings.Repeat(strings.Repeat(" ", 2), int(ctx.Depth)), s.String())
					rl.ResetLine()
				}, &sexpr.SExprVisitCtx{Depth: 0})
				// val, err := s.Eval()
				// if err != nil {
				// 	fmt.Printf("ERROR: could not evaluate sexpr from `%s` %v", line, err)
				// 	rl.ResetLine()
				// }
				// if val != nil {
				// 	prec := int(val.Prec())
				// 	acc := val.Acc()
				// 	if acc == big.Exact {
				// 		fmt.Printf("%s ", acc)
				// 		if val.IsInt() {
				// 			prec = 0
				// 		} else {
				// 			prec = 1
				// 		}
				// 	}
				// 	fmt.Print(val.Text('f', prec))
				// }
			default:
				rl.Put(input.Value())
			}
		}
	}
}
