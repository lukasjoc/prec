package sexpr

import (
	"errors"
	"fmt"
	"math/big"
)

// TODO: sep package

func (s *SExpr) Eval() (*big.Float, error) {
	if s.isAtom() {
		return s.evalAtom()
	} else if s.isNil() {
		return s.evalNil()
	} else if s.isList() {
		return s.evalList()
	}
	return nil, errors.New("eval: invalid type for evaluation. Neither list, nil nor atom")
}

func (s *SExpr) evalAtom() (*big.Float, error) {
	if s.token == nil || !s.token.IsConst() {
		return nil, fmt.Errorf("evalAtom:%d cannot evaluate atom type `%v`", s.token.Offset(), s.String())
	}
	f := new(big.Float)
	_, err := fmt.Sscan(s.token.Value(), f)
	if err != nil {
		return nil, fmt.Errorf("evalAtom: %w", err)
	}
	return f, nil
}

func (s *SExpr) evalNil() (*big.Float, error) { return nil, nil }

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

func (s *SExpr) evalList() (*big.Float, error) {
	if s == nil || s.elements == nil || len(s.elements) == 0 || !s.elements[0].isOp() {
		return nil, nil
	}
	op := s.elements[0]
	flatvals := []big.Float{}
	operands := s.elements[1:]
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
