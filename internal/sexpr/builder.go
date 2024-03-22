package sexpr

import (
	"errors"
	"fmt"

	"github.com/lukasjoc/prec/internal/lex"
)

// errUnterminatedList a list was opened but never closed
var errUnterminatedList = errors.New("unterminated list")

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
		return SExpr{}, fmt.Errorf("list: %w", errUnterminatedList)
	}
	// TODO: fix error with ignored trailing parens in lists
	// that should lead to a ErrUnterminatedList
	if tok2.IsClosePar() {
		b.next()
		return SExpr{sexprTypeNil, nil, nil}, nil
	}
	var tokerr error = nil
	args := []*SExpr{}
	for {
		tok, err := b.peek()
		if err != nil {
			tokerr = fmt.Errorf("list: %w", errUnterminatedList)
			break
		}
		if tok.IsClosePar() {
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
	return SExpr{sexprTypeList, nil, args}, tokerr
}

func (b *Builder) Build() (SExpr, error) {
	tok, err := b.next()
	if err != nil {
		return SExpr{}, fmt.Errorf("build: %w", err)
	}
	if tok.IsConst() || tok.IsOp() {
		return b.atom(tok)
	} else if tok.IsOpenPar() {
		return b.list(tok)
	}
	return SExpr{}, fmt.Errorf("build: invalid entrypoint: `%s`", tok.Value())
}
