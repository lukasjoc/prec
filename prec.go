package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

//go:generate stringer -type=TokType
type TokType int

const (
	Const TokType = iota
	OpenPar
	ClosePar
	Space
	Dunno
	Eof
)

var EOF = errors.New("EOF")
var TODO = errors.New("TODO: not yet implemented")

const EOFCh rune = '\000'

type Tok struct {
	Typ    TokType
	Offset int
	Value  string
}

type Lexer struct {
	source []rune
	left   int
	pos    int
	last   int
}

func NewLexer(source string) Lexer { return Lexer{[]rune(source), len(source), -1, 0} }
func (l *Lexer) peekable() bool    { return l.pos+1 < len(l.source) }
func (l *Lexer) peek() rune {
	if l.peekable() {
		return l.source[l.pos+1]
	}
	return EOFCh
}
func (l *Lexer) eat() rune {
	if l.peekable() {
		l.pos += 1
		l.left -= 1
		return l.source[l.pos]
	}
	return EOFCh
}
func (l *Lexer) eatWhile(pred func(ch rune) bool) {
	for {
		if !pred(l.peek()) {
			break
		}
		l.eat()
	}
}
func (l *Lexer) span(from int, to int) string {
	if len(l.source) == 0 {
		return ""
	}
	return string(l.source)[from:to]
}
func (l Lexer) Peek() (*Tok, error) { return l.Next() }
func (l *Lexer) skipWhile(typ TokType) error {
	var tokerr error = nil
	for {
		tok, err := l.Peek()
		if err != nil {
			tokerr = err
			break
		}
		if tok.Typ != typ {
			break
		}
		_, err = l.Next()
		if err != nil {
			tokerr = err
			break
		}
	}
	return tokerr
}
func (l *Lexer) Next() (*Tok, error) {
	ch := l.eat()
	if ch == EOFCh {
		return nil, EOF
	}
	typ := Dunno
	if ch == '(' {
		typ = OpenPar
	} else if ch == ')' {
		typ = ClosePar
	} else if unicode.IsDigit(ch) {
		// TODO: support for floats
		// TODO: support for neg. numbers
		l.eatWhile(func(ch rune) bool { return unicode.IsDigit(ch) })
		typ = Const
	} else if unicode.IsSpace(ch) {
		l.eatWhile(func(ch rune) bool { return unicode.IsSpace(ch) })
		typ = Space
	}
	toklen := len(l.source) - l.left
	tok := &Tok{typ, l.last, l.span(l.last, toklen)}
	l.last = toklen
	return tok, nil
}

//go:generate stringer -type=SExprType
type SExprType int

const (
	// e.g. 5
	Atom SExprType = iota
	// e.g. (1 2 3) or (+ 1 2)
	List
	// e.g. ()
	Nil
)

type SExpr struct {
	Typ  SExprType
	Tok  *Tok
	Args []*SExpr
}

type SExprVisitCtx struct{ Depth int }

type SExprVisitFunc = func(s *SExpr, ctx *SExprVisitCtx)

func (s *SExpr) visitAtom(f SExprVisitFunc, ctx *SExprVisitCtx) { f(s, ctx) }
func (s *SExpr) visitNil(f SExprVisitFunc, ctx *SExprVisitCtx)  { f(s, ctx) }
func (s *SExpr) visitList(f SExprVisitFunc, ctx *SExprVisitCtx) {
	f(s, ctx)
	if s.Args == nil {
		return
	}
	ctx.Depth += 1
	for _, arg := range s.Args {
		arg.Visit(f, ctx)
	}
	ctx.Depth -= 1
}
func (s *SExpr) Visit(f SExprVisitFunc, ctx *SExprVisitCtx) {
	if s.Typ == Atom {
		s.visitAtom(f, ctx)
	} else if s.Typ == Nil {
		s.visitNil(f, ctx)
	} else if s.Typ == List {
		s.visitList(f, ctx)
	}
}
func (s *SExpr) Eval() (string, error) { return "", TODO }
func (s SExpr) String() string {
	if s.Tok == nil {
		return s.Typ.String()
	}
	return fmt.Sprintf("%s:%s:%s", s.Typ.String(), s.Tok.Typ.String(), s.Tok.Value)
}

type SExprBuilder struct{ lexer Lexer }

func NewSExprBuilder(source string) SExprBuilder {
	lexer := NewLexer(source)
	return SExprBuilder{lexer}
}
func (b *SExprBuilder) peek() (*Tok, error) {
	b.lexer.skipWhile(Space)
	return b.lexer.Peek()
}
func (b *SExprBuilder) next() (*Tok, error) {
	b.lexer.skipWhile(Space)
	return b.lexer.Next()
}
func (b *SExprBuilder) atom(tok *Tok) (SExpr, error) { return SExpr{Atom, tok, nil}, nil }
func (b *SExprBuilder) list(tok *Tok) (SExpr, error) {
	// check if (empty, nil) list
	tok2, _ := b.peek()
	if tok2.Typ == ClosePar {
		b.next()
		return SExpr{Nil, nil, nil}, nil
	}
	var tokerr error = nil
	args := []*SExpr{}
	for {
		tok, err := b.peek()
		if err != nil {
			tokerr = err
			break
		}
		if tok.Typ == ClosePar {
			b.next()
			break
		}
		expr, err := b.Build()
		if err != nil {
			tokerr = err
			break
		}
		args = append(args, &expr)
	}
	return SExpr{List, nil, args}, tokerr
}
func (b *SExprBuilder) Build() (SExpr, error) {
	tok, err := b.next()
	if err != nil {
		return SExpr{}, err
	}
	if tok.Typ == OpenPar {
		return b.list(tok)
	} else if tok.Typ == Const {
		return b.atom(tok)
	}
	return SExpr{}, fmt.Errorf("invalid entry: `%v`", tok.Value)
}

type SExprDumper struct{ Indent int }

func (d *SExprDumper) Stdout(s *SExpr) {
	if s == nil {
		return
	}
	indentString := strings.Repeat(" ", d.Indent)
	s.Visit(func(s *SExpr, ctx *SExprVisitCtx) {
		if s == nil {
			return
		}
		fmt.Printf("%s%s\n", strings.Repeat(indentString, ctx.Depth), s.String())
	}, &SExprVisitCtx{Depth: 0})
}

func main() {
	b := NewSExprBuilder("(1 (0) ((100 100 100) () () (1 2 3 10)))")
	expr, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("expr build failed with %v", err))
	}
	dumper := SExprDumper{Indent: 2}
	dumper.Stdout(&expr)
}
