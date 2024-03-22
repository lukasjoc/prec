package sexpr

import (
	"fmt"

	"github.com/lukasjoc/prec/internal/lex"
)

//go:generate stringer -type=sexprType
type sexprType int

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
