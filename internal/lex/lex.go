package lex

import (
	"fmt"
	"io"
	"unicode"
)

//go:generate stringer -type=tokenType
type tokenType int

const (
	tokenTypeConst tokenType = iota
	tokenTypeOp
	tokenTypeOpenPar
	tokenTypeClosePar
	tokenTypeSpace
	tokenTypeDunno
	tokenTypeEof
)

const nullch byte = 0x000

var supportedOpMap = map[byte]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
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
func (t Token) IsOpenPar() bool  { return t.typ == tokenTypeOpenPar }
func (t Token) IsClosePar() bool { return t.typ == tokenTypeClosePar }

type Lexer struct {
	source []byte
	left   int
	pos    int
	last   int
}

func New(source string) Lexer   { return Lexer{[]byte(source), len(source), -1, 0} }
func (l *Lexer) peekable() bool { return l.pos+1 < len(l.source) }
func (l *Lexer) peek() byte {
	if l.peekable() {
		return l.source[l.pos+1]
	}
	return nullch
}
func (l *Lexer) eat() byte {
	if l.peekable() {
		l.pos += 1
		l.left -= 1
		return l.source[l.pos]
	}
	return nullch
}
func (l *Lexer) eatWhile(pred func(ch byte) bool) {
	for {
		if !pred(l.peek()) {
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

func (l *Lexer) SkipWhile(typ tokenType) error {
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
	l.SkipWhile(tokenTypeSpace)
}

func (l *Lexer) isConst(at byte) bool {
	if unicode.IsDigit(rune(at)) {
		return true
	}
	ch := l.peek()
	return ch != nullch && (at == '-' || at == '+') && unicode.IsDigit(rune(ch))
}

func (l *Lexer) Next() (*Token, error) {
	ch := l.eat()
	if ch == nullch {
		return nil, fmt.Errorf("next: %w", io.EOF)
	}
	typ := tokenTypeDunno
	if ch == '(' {
		typ = tokenTypeOpenPar
	} else if ch == ')' {
		typ = tokenTypeClosePar
	} else if l.isConst(ch) {
		// TODO: support for floats
		l.eatWhile(func(ch byte) bool { return unicode.IsDigit(rune(ch)) })
		typ = tokenTypeConst
	} else if unicode.IsSpace(rune(ch)) {
		l.eatWhile(func(ch byte) bool { return unicode.IsSpace(rune(ch)) })
		typ = tokenTypeSpace
	} else if supportedOpMap[ch] {
		typ = tokenTypeOp
	}
	toklen := len(l.source) - l.left
	tok := &Token{typ, l.last, l.span(l.last, toklen)}
	l.last = toklen
	return tok, nil
}
