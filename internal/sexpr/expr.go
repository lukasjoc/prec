package sexpr

import (
	"fmt"
	"math"

	"github.com/lukasjoc/prec/internal/lex"
)

//go:generate stringer -type=sexprType
type sexprType uint8

const (
	sexprTypeAtom sexprType = iota
	sexprTypeList
	sexprTypeNil
)

type SExpr struct {
	typ      sexprType
	token    *lex.Token
	elements []*SExpr
}

func (s SExpr) String() string {
	if s.token == nil {
		return s.typ.String()
	}
	return fmt.Sprintf("%s:%s:%s", s.typ.String(), s.token.String(), s.token.Value())
}

func (s *SExpr) isAtom() bool { return s.typ == sexprTypeAtom }
func (s *SExpr) isOp() bool   { return s.isAtom() && s.token.IsOp() }
func (s *SExpr) isList() bool { return s.typ == sexprTypeList }
func (s *SExpr) isNil() bool  { return s.typ == sexprTypeNil }

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

type Builder struct{ l lex.Lexer }

func NewBuilder(source string) Builder {
	return Builder{l: lex.New(source)}
}

func (b *Builder) peek() (*lex.Token, error) {
	b.l.SkipWhileSpace()
	return b.l.Peek()
}

func (b *Builder) next() (*lex.Token, error) {
	b.l.SkipWhileSpace()
	return b.l.Next()
}

func (b *Builder) atom(tok *lex.Token) (SExpr, error) {
	return SExpr{sexprTypeAtom, tok, nil}, nil
}

func (b *Builder) list(tok *lex.Token) (SExpr, error) {
	tok2, err := b.peek()
	if err != nil {
		return SExpr{}, fmt.Errorf("list: unterminated list %w", err)
	}
	// TODO: fix error with ignored trailing parens in lists
	// that should lead to a ErrUnterminatedList
	if tok2.IsClosePar() {
		b.next()
		return SExpr{sexprTypeNil, nil, nil}, nil
	}
	var tokerr error = nil
	elements := []*SExpr{}
	for {
		tok, err := b.peek()
		if err != nil {
			tokerr = fmt.Errorf("list: unterminated list %w", err)
			break
		}
		if tok.IsClosePar() {
			b.next()
			break
		}
		sub, err := b.Build()
		if err != nil {
			tokerr = err
			break
		}
		elements = append(elements, &sub)
	}
	return SExpr{sexprTypeList, nil, elements}, tokerr
}

func (b *Builder) Build() (SExpr, error) {
	tok, err := b.next()
	if err != nil {
		return SExpr{}, fmt.Errorf("build: %w", err)
	}
	if tok.IsConst() || tok.IsOp() || tok.IsIdent() {
		return b.atom(tok)
	} else if tok.IsOpenPar() {
		return b.list(tok)
	}
	return SExpr{}, fmt.Errorf("build:%d invalid entrypoint: `%s`", tok.Offset(), tok.Value())
}
