package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"unicode"
)

//go:generate stringer -type=TokType
type TokType int

const (
	Const TokType = iota
	Op
	OpenPar
	ClosePar
	Space
	Dunno
	Eof
)

func isSupportedOp(ch rune) bool { return ch == '+' || ch == '-' || ch == '*' || ch == '/' }

// ErrEOF there are no toks left to read
var ErrEOF = errors.New("EOF")

// ErrTODO the function is not implemented
var ErrTODO = errors.New("TODO: not yet implemented")

// ErrUnterminatedList a list was opened but never closed
var ErrUnterminatedList = errors.New("unterminated list")

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
	return fmt.Errorf("skipWhile: %w", tokerr)
}
func (l *Lexer) Next() (*Tok, error) {
	ch := l.eat()
	if ch == EOFCh {
		return nil, fmt.Errorf("next: %w", ErrEOF)
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
	} else if isSupportedOp(ch) {
		typ = Op
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

func (s SExpr) String() string {
	if s.Tok == nil {
		return s.Typ.String()
	}
	return fmt.Sprintf("%s:%s:%s", s.Typ.String(), s.Tok.Typ.String(), s.Tok.Value)
}

func (s *SExpr) IsAtom() bool { return s.Typ == Atom }
func (s *SExpr) IsOp() bool   { return s.IsAtom() && s.Tok.Typ == Op }
func (s *SExpr) IsList() bool { return s.Typ == List }
func (s *SExpr) IsNil() bool  { return s.Typ == Nil }

type SExprVisitFunc = func(s *SExpr, ctx *SExprVisitCtx)
type SExprVisitCtx struct {
	// TODO: imagine a max depth for savety :)
	Depth int
}

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
	if s.IsAtom() {
		s.visitAtom(f, ctx)
	} else if s.IsNil() {
		s.visitNil(f, ctx)
	} else if s.IsList() {
		s.visitList(f, ctx)
	}
}

func applyOp(op string, valsPtr *[]big.Float) (*big.Float, error) {
	if valsPtr == nil {
		return nil, nil
	}
	vals := *valsPtr
	if len(vals) == 0 {
		return nil, errors.New("applyOp: expected at least a single operand for evaluation")
	}
	x := vals[0]
	for _, y := range vals[1:] {
		switch op {
		case "+":
			x.Add(&x, &y)
		case "-":
			x.Sub(&x, &y)
		case "*":
			x.Mul(&x, &y)
		case "/":
			x.Quo(&x, &y)
		default:
			return nil, errors.New("applyOp: invalid op expected one of `+-*/`")
		}
	}
	return &x, nil
}
func (s *SExpr) Eval() (*big.Float, error) {
	if s.IsAtom() {
		return s.evalAtom()
	} else if s.IsNil() {
		return s.evalNil()
	} else if s.IsList() {
		return s.evalList()
	}
	return nil, errors.New("eval: invalid type for evaluation. Neither list, nil nor atom")
}
func (s *SExpr) evalAtom() (*big.Float, error) {
	if s.Tok == nil || s.Tok.Typ != Const {
		return nil, fmt.Errorf("evalAtom: cannot evaluate atom type `%v`", s.String())
	}
	f := new(big.Float)
	_, err := fmt.Sscan(s.Tok.Value, f)
	if err != nil {
		return nil, fmt.Errorf("evalAtom: %w", err)
	}
	return f, nil
}
func (s *SExpr) evalNil() (*big.Float, error) { return nil, nil }
func (s *SExpr) evalList() (*big.Float, error) {
	if s == nil || s.Args == nil || len(s.Args) == 0 || !s.Args[0].IsOp() {
		return nil, nil
	}
	op := s.Args[0]
	flatvals := []big.Float{}
	operands := s.Args[1:]
	for _, operand := range operands {
		if operand.IsOp() {
			return nil, fmt.Errorf("evalList: invalid operand type `%v`", operand.String())
		}
		opval, err := operand.Eval()
		if err != nil {
			return nil, fmt.Errorf("evalList: %w", err)
		}
		if opval != nil {
			flatvals = append(flatvals, *opval)
		}
	}
	return applyOp(op.Tok.Value, &flatvals)
}

type SExprBuilder struct{ lexer Lexer }

func NewSExprBuilder(source string) SExprBuilder { return SExprBuilder{NewLexer(source)} }
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
	tok2, err := b.peek()
	if err != nil {
		return SExpr{}, fmt.Errorf("list: %w", ErrUnterminatedList)
	}
	// TODO: fix error with ignored trailing parens in lists
	// that should lead to a ErrUnterminatedList
	if tok2.Typ == ClosePar {
		b.next()
		return SExpr{Nil, nil, nil}, nil
	}
	var tokerr error = nil
	args := []*SExpr{}
	for {
		tok, err := b.peek()
		if err != nil {
			tokerr = fmt.Errorf("list: %w", ErrUnterminatedList)
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
		return SExpr{}, fmt.Errorf("build: %w", err)
	}
	if tok.Typ == Const || tok.Typ == Op {
		return b.atom(tok)
	} else if tok.Typ == OpenPar {
		return b.list(tok)
	}
	return SExpr{}, fmt.Errorf("build: invalid entrypoint: `%s`", tok.Value)
}

// TODO: make it bufio.Writer compatible
type SExprDumper struct{ Indent int }

func (d *SExprDumper) StdoutWrite(s *SExpr) {
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
	dumper := SExprDumper{Indent: 2}
	stdin := bufio.NewReader(os.Stdin)
	fmt.Println("prec v1")
	fmt.Println("The precision calculator with the lispy dialect.")
	for {
		fmt.Printf("; ")
		raw, _ := stdin.ReadString('\n')
		line := strings.TrimSpace(raw)
		if len(line) == 0 {
			continue
		}
		builder := NewSExprBuilder(line)
		s, err := builder.Build()
		if err != nil && !errors.Is(err, ErrEOF) {
			fmt.Printf("ERROR: could not build sexpr from `%s` %v\n", line, err)
			continue
		}
		dumper.StdoutWrite(&s)
		val, err := s.Eval()
		if err != nil {
			fmt.Printf("ERROR: could not evaluate sexpr from `%s` %v\n", line, err)
			continue
		}
		if val == nil {
			continue
		}
		prec := int(val.Prec())
		acc := val.Acc()
		if acc == big.Exact {
			fmt.Printf("%s ", acc)
			if val.IsInt() {
				prec = 0
			} else {
				prec = 1
			}
		}
		fmt.Printf("%s\n", val.Text('f', prec))
	}
}
