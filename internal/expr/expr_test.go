package expr

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestExpr(t *testing.T) {
	id, data := loadFixture(t)
	for _, input := range data {
		sum := fmt.Sprintf("%x", sha256.Sum256([]byte(input)))[:8]
		t.Run(fmt.Sprintf("TestExpr-%s-%s", id, sum), func(t *testing.T) {
			ctx := NewEvalCtx()
			p := NewParser([]byte(input))
			expr, parseErr := p.Parse()
			if parseErr != nil {
				snapshotAssert(t, input, parseErr.Error())
				return
			}
			ret, evalErr := ctx.Eval(&expr)
			if evalErr != nil {
				snapshotAssert(t, input, evalErr.Error())
			} else {
				if ret != nil {
					snapshotAssert(t, input, ret.String())
				} else {
					snapshotAssert(t, input, "nil")
				}
			}
		})
	}
}
