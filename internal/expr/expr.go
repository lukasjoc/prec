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
	typ      sexprType
	token    *lex.Token
	elements []*Expr // !Recursive
}

func (e Expr) String() string {
	if e.token == nil {
		return e.typ.String()
	}
	return fmt.Sprintf("%s:%s:%s", e.typ.String(), e.token.String(), e.token.Value())
}

func (e *Expr) isAtom() bool { return e.typ == sexprTypeAtom }
func (e *Expr) isOp() bool   { return e.isAtom() && e.token.IsOp() }
func (e *Expr) isList() bool { return e.typ == sexprTypeList }
func (e *Expr) isNil() bool  { return e.typ == sexprTypeNil }

type ExprVisitFunc = func(e *Expr, ctx *ExprVisitCtx)
type ExprVisitCtx struct {
	Depth uint8
}

func (e *Expr) visitAtom(f ExprVisitFunc, ctx *ExprVisitCtx) { f(e, ctx) }
func (e *Expr) visitNil(f ExprVisitFunc, ctx *ExprVisitCtx)  { f(e, ctx) }
func (e *Expr) visitList(f ExprVisitFunc, ctx *ExprVisitCtx) {
	f(e, ctx)
	if e.elements == nil {
		return
	}
	if ctx.Depth+1 == math.MaxUint8 {
		// TODO: prob should touch the sig here
		panic(fmt.Errorf("depth would exceed limit of visitor: max %d", math.MaxUint8))
	}
	ctx.Depth += 1
	for _, element := range e.elements {
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

func (e *Expr) Eval() (*big.Float, error) {
	if e.isAtom() {
		return e.evalAtom()
	} else if e.isNil() {
		return e.evalNil()
	} else if e.isList() {
		return e.evalList()
	}
	return nil, errors.New("eval: invalid type for evaluation. Neither list, nil nor atom")
}

func (e *Expr) evalAtom() (*big.Float, error) {
	if e.token == nil || !e.token.IsConst() {
		return nil, fmt.Errorf("evalAtom:%d cannot evaluate atom type `%v`", e.token.Offset(), e.String())
	}
	f := new(big.Float)
	_, err := fmt.Sscan(e.token.Value(), f)
	if err != nil {
		return nil, fmt.Errorf("evalAtom: %w", err)
	}
	return f, nil
}

func (e *Expr) evalNil() (*big.Float, error) { return nil, nil }

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

func (e *Expr) evalList() (*big.Float, error) {
	if e == nil || e.elements == nil || len(e.elements) == 0 || !e.elements[0].isOp() {
		return nil, nil
	}
	op := e.elements[0]
	flatvals := []big.Float{}
	operands := e.elements[1:]
	for _, operand := range operands {
		if operand.isOp() {
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
	return applyOp(op.token.Value(), &flatvals)
}
