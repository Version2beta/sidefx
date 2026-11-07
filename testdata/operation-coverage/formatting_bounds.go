package operationcoverage

import "fmt"

// No runtime allocation: this function is only built into static SSA. The
// formatter proof must not allocate a table proportional to source array size.
// Keep the type itself representable on 32-bit targets too.
func LargeFormattingArray() { a := new([1 << 24]any); _ = fmt.Sprint(a[:1]...) }
