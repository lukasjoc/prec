package main

import (
	"fmt"
	"os"

	"github.com/lukasjoc/prec/readline"
)

// func handleInput(line string, dumper sexpr.Dumper) error {
// 	builder := sexpr.NewBuilder(line)
// 	s, err := builder.Build()
// 	if err != nil && !errors.Is(err, io.EOF) {
// 		return fmt.Errorf("ERROR: could not build sexpr from `%s` %v\n", line, err)
// 	}
// 	dumper.Dump(&s)
// 	val, err := s.Eval()
// 	if err != nil {
// 		return fmt.Errorf("ERROR: could not evaluate sexpr from `%s` %v\n", line, err)
// 	}
// 	if val == nil {
// 		return errors.New("invalid value")
// 	}
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
// 	fmt.Printf("%s\n", val.Text('f', prec))
//
// 	return nil
// }

func main() {
	canon, _ := readline.GetTerminalMode(os.Stdin)
	readline.SetTerminalRawMode(os.Stdin, canon)
	defer readline.ResetTerminalRawMode(os.Stdin, canon)

	rl := readline.New(os.Stdin)
	for {
		input := rl.Poll()
		switch input.Key {
		case readline.InputEventGeneric:
			if input.Bytes == nil {
				continue
			}
			fmt.Print(string(*input.Bytes))
			break
		case readline.InputEventArrrowLeft:
			fmt.Print("<")
			break
		case readline.InputEventArrowRight:
			fmt.Print(">")
			break
		}
	}
}
