package test

import (
	"encoding/json/v2"
	"testing"

	"github.com/lukasjoc/prec/internal/expr"
	"github.com/lukasjoc/prec/snap"
)

func TestParse(t *testing.T) {
	testdata, err := loadTest()
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range testdata {
		name := makeTestname(t, input)
		t.Run(name, func(t *testing.T) {
			p := expr.NewParser([]byte(input))
			expr, err := p.Parse()

			var output string

			if err != nil {
				output = err.Error()
			} else {
				out, err := json.Marshal(expr)
				if err != nil {
					t.Fatal(err)
				}
				output = string(out)
			}

			snap.New(t).WithDescription(input).Assert(output)
		})
	}
}
