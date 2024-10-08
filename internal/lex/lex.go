package lex

import (
	"fmt"
	"io"
	"unicode"
)

//go:generate stringer -type=tokenType
type tokenType uint8

const (
	tokenTypeConst tokenType = iota
	tokenTypeOp
	tokenTypeIdent
	tokenTypeOpenPar
	tokenTypeClosePar
	tokenTypeSpace
	tokenTypeInvalid
	tokenTypeEof
)

const eof byte = 0xFF // 255

func isSupportedOp(ch byte) bool {
	switch ch {
	case '+', '-', '*', '/':
		return true
	default:
		return false
	}
}

type Token struct {
	typ    tokenType
	offset int
	value  []byte
}

func (t Token) Value() string    { return string(t.value) }
func (t Token) Bytes() []byte    { return t.value }
func (t Token) String() string   { return t.typ.String() }
func (t Token) Offset() int      { return t.offset }
func (t Token) IsConst() bool    { return t.typ == tokenTypeConst }
func (t Token) IsOp() bool       { return t.typ == tokenTypeOp }
func (t Token) IsIdent() bool    { return t.typ == tokenTypeIdent }
func (t Token) IsOpenPar() bool  { return t.typ == tokenTypeOpenPar }
func (t Token) IsClosePar() bool { return t.typ == tokenTypeClosePar }

type Lexer struct {
	source []byte
	left   int
	pos    int
	last   int
}

// TODO: should accept io.Reader and `stream` the tokens
func New(source []byte) *Lexer  { return &Lexer{source, len(source), -1, 0} }
func (l *Lexer) peekable() bool { return l.pos+1 < len(l.source) }
func (l *Lexer) peek() byte {
	if l.peekable() {
		return l.source[l.pos+1]
	}
	return eof
}
func (l *Lexer) eat() byte {
	if l.peekable() {
		l.pos++
		l.left--
		return l.source[l.pos]
	}
	return eof
}
func (l *Lexer) eatWhile(pred func(ch byte) bool) {
	for {
		ch := l.peek()
		if ch == eof || !pred(ch) {
			break
		}
		l.eat()
	}
}
func (l *Lexer) span(from int, to int) []byte {
	if len(l.source) == 0 {
		return nil
	}
	return []byte(l.source)[from:to]
}

// FIXME: Shady bizz
func (l Lexer) Peek() (*Token, error) { return l.Next() }

func (l *Lexer) skipWhile(typ tokenType) error {
	var tokerr error = nil
	for {
		tok, err := l.Peek()
		if err != nil {
			tokerr = err
			break
		}
		if tok.typ != typ {
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

func (l *Lexer) SkipWhileSpace() {
	l.skipWhile(tokenTypeSpace)
}

func (l *Lexer) isConst(at byte) bool {
	if unicode.IsDigit(rune(at)) {
		return true
	}
	ch := l.peek()
	return ch != eof && (at == '-' || at == '+') && unicode.IsDigit(rune(ch))
}

func canStartIdent(ch byte) bool {
	//FIXME: Why is this not inlined?
	return unicode.IsLetter(rune(ch))
}

func isIdentChar(ch byte) bool {
	if ch == '-' || ch == '\'' || unicode.IsLetter(rune(ch)) || unicode.IsNumber(rune(ch)) {
		return true
	}
	return false
}

func (l *Lexer) tryParseIdent(at byte, typ *tokenType) {
	if !canStartIdent(at) {
		panic(fmt.Errorf("bad assumptions:%d char `%v` cannot start an ident", l.pos, at))
	}
	start := l.pos
	leftStart := l.left
	l.eatWhile(func(ch byte) bool { return isIdentChar(ch) })
	ch := l.peek()
	if ch != eof && ch != '(' && ch != ')' && !unicode.IsSpace(rune(ch)) && !isIdentChar(ch) {
		l.pos = start
		l.left = leftStart
		return
	}
	*typ = tokenTypeIdent
}

func (l *Lexer) tryParseConst(at byte, typ *tokenType) {
	if !l.isConst(at) {
		return
	}
	// TODO: support for floats
	start := l.pos
	leftStart := l.left
	l.eatWhile(func(ch byte) bool { return unicode.IsDigit(rune(ch)) })
	// check if next char after consume is valid {space | brace | eof}. If it's
	// a letter for example that means that it parsed the number from sth like
	// `42069f` which is not valid. (Obvsly)
	ch := l.peek()
	if ch != eof && ch != '(' && ch != ')' && !unicode.IsSpace(rune(ch)) {
		l.pos = start
		l.left = leftStart
		return
	}
	*typ = tokenTypeConst
}

func (l *Lexer) Next() (*Token, error) {
	ch := l.eat()
	if ch == eof {
		return nil, fmt.Errorf("next: %w", io.EOF)
	}
	typ := tokenTypeInvalid
	if ch == '(' {
		typ = tokenTypeOpenPar
	} else if ch == ')' {
		typ = tokenTypeClosePar
	} else {
		if canStartIdent(ch) {
			l.tryParseIdent(ch, &typ)
		} else if l.isConst(ch) {
			l.tryParseConst(ch, &typ)
		} else if unicode.IsSpace(rune(ch)) {
			l.eatWhile(func(ch byte) bool { return unicode.IsSpace(rune(ch)) })
			typ = tokenTypeSpace
		} else if isSupportedOp(ch) {
			typ = tokenTypeOp
		}
	}
	toklen := len(l.source) - l.left
	if typ == tokenTypeInvalid {
		return nil, fmt.Errorf("next:%d bad syntax `%s`", l.pos, string(l.source[l.pos]))
	}
	tok := &Token{typ, l.last, l.span(l.last, toklen)}
	l.last = toklen
	return tok, nil
}
