package main

import (
	"errors"
	"fmt"
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

const EOFCh rune = '\000'

type Tok struct {
	typ    TokType
	offset int
	len    int
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
		// TODO: float support
		l.eatWhile(func(ch rune) bool { return unicode.IsDigit(ch) })
		typ = Const
	} else if unicode.IsSpace(ch) {
		l.eatWhile(func(ch rune) bool { return unicode.IsSpace(ch) })
		typ = Space
	}
	toklen := len(l.source) - l.left
	tok := &Tok{typ, l.last, toklen}
	l.last = toklen
	return tok, nil
}

// TODO: skip spaces func

func (l *Lexer) Span(tok *Tok) (string, error) {
	sourceLen := len(l.source)
	if sourceLen == 0 {
		return "", EOF
	}
	return string(l.source)[tok.offset:tok.len], nil
}

func main() {
	source := "(+ 10000 (* 10 (/ 100 100)))"
	println("SOURCE: ", source)
	lexer := NewLexer(source)
	for {
		tok, err := lexer.Next()
		if err != nil {
			fmt.Println("INFO: EOF (Caught OK)")
			break
		}
		span, err := lexer.Span(tok)
		fmt.Printf("TOK: %#v\t(%4d..%4d)\t%#v\n", tok.typ.String(), tok.offset, tok.len, span)
	}
}
