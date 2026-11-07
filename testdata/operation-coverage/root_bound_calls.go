package operationcoverage

import "os"

func RootBoundWrite(r *os.Root)              { f := r.OpenFile; _, _ = f("never", os.O_WRONLY|os.O_CREATE, 0600) }
func RootBoundDynamic(r *os.Root, flags int) { f := r.OpenFile; _, _ = f("never", flags, 0600) }
func RootBoundReadMode(r *os.Root)           { f := r.OpenFile; _, _ = f("never", os.O_RDONLY, 0601) }
func RootExpressionWrite(r *os.Root) {
	f := (*os.Root).OpenFile
	_, _ = f(r, "never", os.O_WRONLY|os.O_CREATE, 0600)
}
func RootExpressionDynamic(r *os.Root, flags int) {
	f := (*os.Root).OpenFile
	_, _ = f(r, "never", flags, 0600)
}
func RootExpressionReadMode(r *os.Root) {
	f := (*os.Root).OpenFile
	_, _ = f(r, "never", os.O_RDONLY, 0601)
}
func RootBoundCaller(r *os.Root) { RootBoundWrite(r) }
func RootBoundAlternatives(r, s *os.Root, choose bool) {
	f := r.OpenFile
	if choose {
		f = s.OpenFile
	}
	_, _ = f("never", os.O_WRONLY|os.O_CREATE, 0600)
}
