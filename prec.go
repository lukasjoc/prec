package main

import (
	// "errors"
	// "fmt"
	// "io"
	// "math/big"
	"os"

	"github.com/lukasjoc/prec/internal/readline"
	"github.com/lukasjoc/prec/internal/sexpr"
)

func main() {
	canon, _ := readline.GetTerminalMode(os.Stdin)
	readline.SetTerminalRawMode(os.Stdin, canon)
	defer readline.ResetTerminalRawMode(os.Stdin, canon)

	dumper := sexpr.NewDumper()
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
			case 'q', 'Q':
				break outer
			case '\r', '\n':
				rl.MoveToNextLine()
				line := rl.Text()
				builder := sexpr.NewBuilder(line)
				s, _ := builder.Build()
				// if err != nil && !errors.Is(err, io.EOF) {
				// 	fmt.Printf("ERROR: could not build sexpr from `%s` %v", line, err)
				// }
				dumper.Dump(&s)
				// val, err := s.Eval()
				// if err != nil {
				// 	fmt.Printf("ERROR: could not evaluate sexpr from `%s` %v", line, err)
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
				// rl.MoveToNextLine()
				rl.ClearLine()
			default:
				rl.Put(input.Value())
			}
		}
	}
}
