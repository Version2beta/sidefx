package effects

import (
	"math/bits"
	"testing"
)

func TestSetCapacityAndUniqueNames(t *testing.T) {
	// All positions are occupied, including internal Unknown. A future addition
	// must not silently reuse a bit or label, or expose Unknown as an effect.
	if ControlPanic != Set(1)<<63 || len(names) != 63 {
		t.Fatalf("expected full uint64 vocabulary: last=%x names=%d", ControlPanic, len(names))
	}
	all := Unknown
	seen := map[string]bool{}
	for bit, name := range names {
		if bits.OnesCount64(uint64(bit)) != 1 || bit&Unknown != 0 || name == "" || seen[name] {
			t.Fatalf("invalid/duplicate vocabulary %x %q", bit, name)
		}
		all |= bit
		seen[name] = true
	}
	if all != ^Set(0) || len(all.Names()) != 63 || len(Unknown.Names()) != 0 {
		t.Fatalf("missing/unknown output bits: %x", all)
	}
}
