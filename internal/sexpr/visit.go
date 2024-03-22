package sexpr

type sexprVisitFunc = func(s *SExpr, ctx *sexprVisitCtx)
type sexprVisitCtx struct {
	// TODO: imagine a max depth for savety :)
	depth int
}

func (s *SExpr) visitAtom(f sexprVisitFunc, ctx *sexprVisitCtx) { f(s, ctx) }
func (s *SExpr) visitNil(f sexprVisitFunc, ctx *sexprVisitCtx)  { f(s, ctx) }
func (s *SExpr) visitList(f sexprVisitFunc, ctx *sexprVisitCtx) {
	f(s, ctx)
	if s.elements == nil {
		return
	}
	ctx.depth += 1
	for _, element := range s.elements {
		element.Visit(f, ctx)
	}
	ctx.depth -= 1
}

func (s *SExpr) Visit(f sexprVisitFunc, ctx *sexprVisitCtx) {
	if s.isAtom() {
		s.visitAtom(f, ctx)
	} else if s.isNil() {
		s.visitNil(f, ctx)
	} else if s.isList() {
		s.visitList(f, ctx)
	}
}
