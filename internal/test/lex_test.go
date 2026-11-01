package test

import (
	"encoding/json/v2"
	"testing"

	"github.com/lukasjoc/prec/internal/lex"
	"github.com/lukasjoc/prec/snap"
)

func TestLex(t *testing.T) {
	testdata, err := loadTest()
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range testdata {
		name := makeTestname(t, input)
		t.Run(name, func(t *testing.T) {
			l := lex.New([]byte(input))

			output := []string{}

			for l.Peekable() {
				tok, err := l.Next()
				if err != nil {
					output = append(output, err.Error())
				} else {
					output = append(output, tok.Type.String())
				}
			}

			out, err := json.Marshal(output)
			if err != nil {
				t.Fatal(err)
			}

			snap.New(t).WithDescription(input).Assert(string(out))
		})
	}
}
