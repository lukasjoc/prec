package expr

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/lukasjoc/prec/internal/lex"
)

//go:generate stringer -type=sexprType
type sexprType uint8

const (
	sexprTypeAtom sexprType = iota
	sexprTypeList
	sexprTypeNil
)

type Expr struct {
	Type  sexprType
	Token *lex.Token
	Elems []*Expr
}

func (e Expr) String() string {
	if e.Token == nil {
		return e.Type.String()
	}
	return fmt.Sprintf("%s:%s:%s", e.Type.String(), e.Token.String(), e.Token.Value())
}

func (e *Expr) isAtom() bool { return e.Type == sexprTypeAtom }
func (e *Expr) isOp() bool   { return e.isAtom() && e.Token.IsOp() }
func (e *Expr) isList() bool { return e.Type == sexprTypeList && len(e.Elems) != 0 }
func (e *Expr) isNil() bool  { return e.Type == sexprTypeNil }

func (e *Expr) Head() *Expr {
	if !e.isList() {
		return nil
	}
	return e.Elems[0]
}

func (e *Expr) Tail() []*Expr {
	if !e.isList() {
		return nil
	}
	return e.Elems[1:]
}

type ExprVisitFunc = func(e *Expr, ctx *ExprVisitCtx)
type ExprVisitCtx struct {
	Depth uint8
}

func (e *Expr) visitAtom(f ExprVisitFunc, ctx *ExprVisitCtx) { f(e, ctx) }
func (e *Expr) visitNil(f ExprVisitFunc, ctx *ExprVisitCtx)  { f(e, ctx) }
func (e *Expr) visitList(f ExprVisitFunc, ctx *ExprVisitCtx) {
	f(e, ctx)
	if ctx.Depth+1 == math.MaxUint8 {
		// TODO: prob should touch the sig here
		panic(fmt.Errorf("depth would exceed limit of visitor: max %d", math.MaxUint8))
	}
	ctx.Depth += 1
	for _, element := range e.Elems {
		element.Visit(f, ctx)
	}
	ctx.Depth -= 1
}

func (e *Expr) Visit(f ExprVisitFunc, ctx *ExprVisitCtx) {
	if e.isAtom() {
		e.visitAtom(f, ctx)
	} else if e.isNil() {
		e.visitNil(f, ctx)
	} else if e.isList() {
		e.visitList(f, ctx)
	}
}

func (e *Expr) Eval(ctx *evalCtx) (*big.Float, error) {
	if e.isAtom() {
		return e.evalAtom()
	} else if e.isNil() {
		return e.evalNil()
	} else if e.isList() {
		return e.evalList(ctx)
	}
	return nil, errors.New("eval: invalid type for evaluation. Neither list, nil nor atom")
}

func (e *Expr) evalAtom() (*big.Float, error) {
	if e.Token == nil || !e.Token.IsConst() {
		return nil, fmt.Errorf("evalAtom:%d cannot evaluate atom type `%v`", e.Token.Offset(), e.String())
	}
	f := new(big.Float)
	_, err := fmt.Sscan(e.Token.Value(), f)
	if err != nil {
		return nil, fmt.Errorf("evalAtom: %w", err)
	}
	return f, nil
}

func (e *Expr) evalNil() (*big.Float, error) { return nil, nil }

func (e *Expr) evalList(ctx *evalCtx) (*big.Float, error) {
	head := e.Head()
	if !head.isAtom() || head.isAtom() && head.Token.IsConst() {
		// NOTE: accept that it's probably some sort of nested list etc.. that we
		// just ignore for eval (for now).
		return nil, nil
	}
	name := head.Token.Value()
	if !ctx.Prelude.Defined(name) {
		return nil, fmt.Errorf("evalList: prelude not defined for: %v", name)
	}

	// Recursively resolve the rest of the list elements as much as possible.
	elems := []*big.Float{}
	for _, elem := range e.Tail() {
		// TODO: make sure to return an error if we try to apply a list to a
		// prelude at the very end.. like (min (5 4))
		if elem.isOp() || elem.isNil() {
			return nil, fmt.Errorf("evalList: invalid operand type `%v`", elem.String())
		}
		opval, err := ctx.Eval(elem)
		if err != nil {
			return nil, fmt.Errorf("evalList: %w", err)
		}
		if opval != nil {
			elems = append(elems, opval)
		}
	}
	return ctx.Prelude.ApplyUnchecked(name, elems)
}
