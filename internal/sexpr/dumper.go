package sexpr

import (
	"fmt"
	"strings"
)

// TODO: make it bufio.Writer compatible
type Dumper struct{ Indent int }

func NewDumper() Dumper {
	return Dumper{Indent: 2}
}

func (d *Dumper) Dump(s *SExpr) {
	if s == nil {
		return
	}
	indentString := strings.Repeat(" ", d.Indent)
	s.Visit(func(s *SExpr, ctx *sexprVisitCtx) {
		if s == nil {
			return
		}
		fmt.Printf("%s%s\n", strings.Repeat(indentString, ctx.depth), s.String())
	}, &sexprVisitCtx{depth: 0})
}
