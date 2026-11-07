package effects

import (
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/ssa"
)

func TestOpenFileMalformedArityIsUncertain(t *testing.T) {
	pkg := types.NewPackage("os", "os")
	root := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Root", nil), types.NewStruct(nil, nil), nil)
	recv := types.NewVar(token.NoPos, pkg, "r", types.NewPointer(root))
	for _, c := range []struct {
		recv    *types.Var
		lengths []int
	}{{nil, []int{0, 1, 2, 4, 5}}, {recv, []int{0, 1, 2, 5}}} {
		fn := types.NewFunc(token.NoPos, pkg, "OpenFile", types.NewSignatureType(c.recv, nil, nil, types.NewTuple(), types.NewTuple(), false))
		for _, n := range c.lengths {
			got, ok := LookupCall(fn, make([]ssa.Value, n))
			if !ok || got != FSRead|FSWrite|FSCreate|ResourceAcquire|Unknown {
				t.Errorf("%s arity=%d got=%x known=%t", Key(fn), n, got, ok)
			}
		}
	}
}
