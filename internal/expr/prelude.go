package expr

import (
	"errors"
	"fmt"
	"math/big"
)

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

var errWrongNumberArgs = errors.New("wrong number of arguments")

func newPrelude() *prelude {
	p := &prelude{}

	p.register("+", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`+` %w", errWrongNumberArgs)
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
			return nil, fmt.Errorf("prelude:apply:`-` %w", errWrongNumberArgs)
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
			return nil, fmt.Errorf("prelude:apply:`*` %w", errWrongNumberArgs)
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
			return nil, fmt.Errorf("prelude:apply:`/` %w", errWrongNumberArgs)
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
			return nil, fmt.Errorf("prelude:apply:`max` %w", errWrongNumberArgs)
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
			return nil, fmt.Errorf("prelude:apply:`min` %w", errWrongNumberArgs)
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

	p.register("sqrt", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) != 1 {
			return nil, fmt.Errorf("prelude:apply:`sqrt` %w", errWrongNumberArgs)
		}
		z := new(big.Float).Sqrt(elems[0])
		return z, nil
	})

	p.register("neg", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) != 1 {
			return nil, fmt.Errorf("prelude:apply:`neg` %w", errWrongNumberArgs)
		}
		z := new(big.Float).Neg(elems[0])
		return z, nil
	})

	p.register("abs", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) != 1 {
			return nil, fmt.Errorf("prelude:apply:`abs` %w", errWrongNumberArgs)
		}
		z := new(big.Float).Abs(elems[0])
		return z, nil
	})

	p.register("count", func(elems []*big.Float) (*big.Float, error) {
		if len(elems) == 0 {
			return nil, fmt.Errorf("prelude:apply:`count` %w", errWrongNumberArgs)
		}
        // TODO: this shows that we should really resolve to expressions
        // instead of big Float. as that could be any value (atom) or even a list
        // for later..
		return big.NewFloat(float64(len(elems))), nil
	})

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
func (ctx *evalCtx) Eval(e *Expr) (*big.Float, error) {
	return e.Eval(ctx)
}
