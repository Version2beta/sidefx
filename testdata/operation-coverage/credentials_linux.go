//go:build linux

package operationcoverage

import "syscall"

// Only statically analyzed; these functions must never execute in tests.
func SavedCredentials() { _ = syscall.Setresuid(1, 1, 1); _ = syscall.Setresgid(1, 1, 1) }
