package expr

import (
	"errors"
	"fmt"
	"math/big"
)

// FIXME: probaly want to change the return signature of the prelude
// function to return an any otherwise i'd have to create big Floats despite
// needing them..
type preludeFunc = func([]*big.Float) (*big.Float, error)
type prelude struct {
	registered map[string]preludeFunc
}

func (p *prelude) Defined(name string) bool {
	if _, ok := p.registered[name]; ok {
		return true
	}
	return false
}

func (p *prelude) ApplyUnchecked(name string, elems []*big.Float) (*big.Float, error) {
	return p.registered[name](elems)
}

func (p *prelude) register(name string, f preludeFunc) {
	if p.registered == nil {
		p.registered = map[string]preludeFunc{}
	}
	p.registered[name] = f
}

var errMissingElems = errors.New("missing operands for evaluation")

func newPrelude() *prelude {
	p := &prelude{}

	p.register("+", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`+` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			acc.Add(acc, lhs)
		}
		return acc, nil
	})

	p.register("-", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`-` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			acc.Sub(acc, lhs)
		}
		return acc, nil
	})

	p.register("*", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`*` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			acc.Mul(acc, lhs)
		}
		return acc, nil
	})

	p.register("/", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`/` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			acc.Quo(acc, lhs)
		}
		return acc, nil
	})

	p.register("max", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`max` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			res := acc.Cmp(lhs)
			if res == -1 {
				acc = lhs
			}
		}
		return acc, nil
	})

	p.register("min", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`min` %w", errMissingElems)
		}
		acc := elems[0]
		rest := elems[1:]
		for _, lhs := range rest {
			res := acc.Cmp(lhs)
			if res == 1 {
				acc = lhs
			}
		}
		return acc, nil
	})

	// TODO: more functions

	return p
}

type evalCtx struct {
	Prelude *prelude
}

func NewEvalCtx() *evalCtx {
	p := newPrelude()
	return &evalCtx{p}
}

// TODO: maybe we just call the Eval directly and pass in the ctx from outside
// This feels a bit overengineered.
func (ctx *evalCtx) With(e *Expr) (*big.Float, error) {
	return e.Eval(ctx)
}
