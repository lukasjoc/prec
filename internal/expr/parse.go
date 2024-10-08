package expr

import (
	"fmt"

	"github.com/lukasjoc/prec/internal/lex"
)

type parser struct{ l *lex.Lexer }

// TODO: should accept io.Reader and `stream` the tokens
func NewParser(source []byte) *parser {
	return &parser{l: lex.New(source)}
}

func (p *parser) peek() (*lex.Token, error) {
	p.l.SkipWhileSpace()
	return p.l.Peek()
}

func (p *parser) next() (*lex.Token, error) {
	p.l.SkipWhileSpace()
	return p.l.Next()
}

func (p *parser) atom(tok *lex.Token) (Expr, error) {
	return Expr{sexprTypeAtom, tok, nil}, nil
}

func (p *parser) list(tok *lex.Token) (Expr, error) {
	tok2, err := p.peek()
	if err != nil {
		return Expr{}, err
	}
	// TODO: fix error with ignored trailing parens in lists
	// that should lead to a ErrUnterminatedList
	if tok2.IsClosePar() {
		p.next()
		return Expr{sexprTypeNil, nil, nil}, nil
	}
	var tokerr error = nil
	elems := []*Expr{}
	for {
		tok, err := p.peek()
		if err != nil {
			tokerr = err
			break
		}
		if tok.IsClosePar() {
			p.next()
			break
		}
		sub, err := p.Parse()
		if err != nil {
			tokerr = err
			break
		}
		elems = append(elems, &sub)
	}
	return Expr{sexprTypeList, nil, elems}, tokerr
}

func (p *parser) Parse() (Expr, error) {
	tok, err := p.next()
	if err != nil {
		return Expr{}, fmt.Errorf("parse: %w", err)
	}
	if tok.IsConst() || tok.IsOp() || tok.IsIdent() {
		return p.atom(tok)
	} else if tok.IsOpenPar() {
		return p.list(tok)
	}
	return Expr{}, fmt.Errorf("parse:%d invalid entrypoint: `%s`", tok.Offset(), tok.Value())
}
