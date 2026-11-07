package main

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/Version2beta/sidefx/internal/analyzer"
)

// Cross-platform package loading, not cross-platform binary/fixture execution.
func TestCredentialOperationsPlatformQualified(t *testing.T) {
	dir := operationCoverageFixture(t)
	for _, target := range []struct {
		os   string
		apis []string
	}{
		{"linux", []string{"syscall.Setuid", "syscall.Setgid", "syscall.Setgroups", "syscall.Setreuid", "syscall.Setregid", "syscall.Setresuid", "syscall.Setresgid"}},
		{"windows", nil},
	} {
		t.Run(target.os, func(t *testing.T) {
			out, diagnostic, status := invoke(t, dir, []string{"GOOS=" + target.os, "GOARCH=amd64", "CGO_ENABLED=0"}, "-json", "-effect", "process.identity", "-max-depth", "0")
			if status != 0 || diagnostic != "" {
				t.Fatal(status, diagnostic)
			}
			var r analyzer.EffectReport
			if err := json.Unmarshal([]byte(out), &r); err != nil {
				t.Fatal(err)
			}
			if len(r.Sites) != len(target.apis) {
				t.Fatalf("platform operation count: %+v", r.Sites)
			}
			for _, s := range r.Sites {
				if !slices.Contains(target.apis, s.API) || !s.Uncertain || !slices.Contains(s.Reasons, "state_condition") || !slices.Equal(s.Effects, []string{"process.identity"}) || len(s.PossibleEffects) > 0 {
					t.Error(s)
				}
			}
		})
	}
}
