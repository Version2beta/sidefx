package operationcoverage

import (
	"context"
	"io/fs"
	"time"
)

func ContextTimeout(c context.Context, d time.Duration) { _, _ = context.WithTimeout(c, d) }
func ContextTimeoutCause(c context.Context)             { _, _ = context.WithTimeoutCause(c, time.Second, nil) }
func ContextDeadline(c context.Context, d time.Time)    { _, _ = context.WithDeadline(c, d) }
func ContextDeadlineCause(c context.Context, d time.Time) {
	_, _ = context.WithDeadlineCause(c, d, nil)
}
func ContextExpired(c context.Context) {
	_, _ = context.WithTimeout(c, 0)
	_, _ = context.WithTimeoutCause(c, -1, nil)
}
func ContextNil(d time.Time) { _, _ = context.WithTimeout(nil, 1); _, _ = context.WithDeadline(nil, d) }
func ContextCancelOnly(c context.Context) {
	_, cancel := context.WithCancel(c)
	cancel()
	_ = context.Background()
	_ = context.TODO()
}
func ContextTimerCaller(c context.Context, d time.Duration) { ContextTimeout(c, d) }
func FSGlob(f fs.GlobFS)                                    { _, _ = f.Glob("*") }
