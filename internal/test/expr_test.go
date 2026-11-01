package test

import (
	"testing"

	"github.com/lukasjoc/prec/internal/expr"
	"github.com/lukasjoc/prec/snap"
)

func TestExpr(t *testing.T) {
	testdata, err := loadTest()
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range testdata {
		name := makeTestname(t, input)
		t.Run(name, func(t *testing.T) {
			ctx := expr.NewEvalCtx()
			p := expr.NewParser([]byte(input))
			expr, err := p.Parse()

			var output string

			if err != nil {
				output = err.Error()
			} else {
				ret, err := ctx.Eval(&expr)
				if err != nil {
					output = err.Error()
				} else {
					if ret != nil {
						output = ret.String()
					}
				}
			}

			snap.New(t).WithDescription(input).Assert(output)
		})
	}
}
