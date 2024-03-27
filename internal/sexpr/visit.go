package sexpr

import (
	"fmt"
	"math"
)

type SExprVisitFunc = func(s *SExpr, ctx *SExprVisitCtx)
type SExprVisitCtx struct {
	Depth uint8
}

func (s *SExpr) visitAtom(f SExprVisitFunc, ctx *SExprVisitCtx) { f(s, ctx) }
func (s *SExpr) visitNil(f SExprVisitFunc, ctx *SExprVisitCtx)  { f(s, ctx) }
func (s *SExpr) visitList(f SExprVisitFunc, ctx *SExprVisitCtx) {
	f(s, ctx)
	if s.elements == nil {
		return
	}
	if ctx.Depth+1 == math.MaxUint8 {
		// TODO: prob should touch the sig here
		panic(fmt.Errorf("depth would exceed limit of visitor: max %d", math.MaxUint8))
	}
	ctx.Depth += 1
	for _, element := range s.elements {
		element.Visit(f, ctx)
	}
	ctx.Depth -= 1
}

func (s *SExpr) Visit(f SExprVisitFunc, ctx *SExprVisitCtx) {
	if s.isAtom() {
		s.visitAtom(f, ctx)
	} else if s.isNil() {
		s.visitNil(f, ctx)
	} else if s.isList() {
		s.visitList(f, ctx)
	}
}
