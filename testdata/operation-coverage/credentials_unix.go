//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package operationcoverage

import "syscall"

// Credentials is analyzed only: never execute credential-changing operations.
func Credentials() {
	_ = syscall.Setuid(1)
	_ = syscall.Setgid(1)
	_ = syscall.Setgroups([]int{1})
	_ = syscall.Setreuid(1, 1)
	_ = syscall.Setregid(1, 1)
}
