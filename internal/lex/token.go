package lex

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

const nullch byte = '\000'

var supportedOpMap = map[byte]bool{
	'+': true,
	'-': true,
	'*': true,
	'/': true,
}

type Token struct {
	typ    tokenType
	offset int
	value  string
}

func (t *Token) Value() string    { return t.value }
func (t *Token) String() string   { return t.typ.String() }
func (t *Token) IsConst() bool    { return t.typ == tokenTypeConst }
func (t *Token) IsOp() bool       { return t.typ == tokenTypeOp }
func (t *Token) IsOpenPar() bool  { return t.typ == tokenTypeOpenPar }
func (t *Token) IsClosePar() bool { return t.typ == tokenTypeClosePar }
