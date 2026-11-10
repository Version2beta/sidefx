package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRecordedGoProxy(t *testing.T) {
	for _, value := range []string{"", "off", "direct", "https://proxy.invalid", "https://proxy.invalid/cache,direct", "https://one.invalid|https://two.invalid", "proxy.invalid/cache", "file:///local/cache", "https://proxy.invalid/path@version"} {
		if got := recordedGoProxy(value); got != value {
			t.Errorf("harmless proxy setting changed: %q", got)
		}
	}
	for _, value := range []string{"https://dummy-user:dummy-password@proxy.invalid", "https://dummy-token@proxy.invalid", "proxy.invalid?dummy-token=not-a-secret", "https://proxy.invalid?", "https://proxy.invalid/cache?dummy-token=not-a-secret,direct", "https://safe.invalid|https://dummy-user:dummy-password@proxy.invalid", "dummy-user:dummy-password@proxy.invalid", "https://dummy-user:%zz@proxy.invalid"} {
		if got := recordedGoProxy(value); got != "[REDACTED proxy URL credentials/query]" {
			t.Error("proxy URL credentials/query not redacted")
		}
	}
}

func TestAcquisitionDoesNotRecordProxyCredentials(t *testing.T) {
	dir := copyFixture(t) // stdlib-only cached fixture; no module fetch is needed.
	evidence := filepath.Join(t.TempDir(), "fresh")
	dummyProxy := "https://dummy-user:dummy-password@proxy.invalid/cache?dummy-token=not-a-secret"
	_, diag, status := invoke(t, dir, []string{"GOPROXY=" + dummyProxy, "GONOPROXY=none", "GOSUMDB=off", "GOTOOLCHAIN=local"}, "acquire", "-out", evidence, "-direct", "./...")
	if status != 0 {
		t.Fatal(status, diag)
	}
	for _, name := range []string{"acquisition-start.json", "acquisition.json", "index.json"} {
		data, err := os.ReadFile(filepath.Join(evidence, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "dummy-user") || strings.Contains(string(data), "dummy-token") || strings.Contains(string(data), "dummy-password") {
			t.Fatal("credential-bearing proxy was recorded in", name)
		}
		if !strings.Contains(string(data), `"GOPROXY":"[REDACTED proxy URL credentials/query]"`) {
			t.Fatal("missing explicit proxy redaction", name)
		}
	}
}

func TestInstalledIndexOfflineGuidanceBeforeLoading(t *testing.T) {
	home := t.TempDir()
	// Keep both os.UserHomeDir platform selectors inside this test's home.
	if _, diag, status := invoke(t, fixtureSource, []string{"HOME=" + home, "USERPROFILE=" + home}, "skill", "install"); status != 0 {
		t.Fatal(diag)
	}
	installed, err := os.ReadFile(filepath.Join(home, ".agents", "skills", "sidefx", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	// Markdown fences are independent of checkout line endings; installed bytes stay unchanged.
	text := strings.ReplaceAll(string(installed), "\r\n", "\n")
	start := strings.Index(text, "```sh\n")
	if start < 0 {
		t.Fatal("missing offline setup block")
	}
	start += len("```sh\n")
	end := strings.Index(text[start:], "\n```")
	if end < 0 {
		t.Fatal("unterminated setup block")
	}
	setup := text[start : start+end]
	for _, required := range []string{"GOPROXY=off", "GONOPROXY=none", "GOSUMDB=off", "GOTOOLCHAIN=local", "GOTELEMETRY=off", `GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly"`} {
		if !strings.Contains(setup, required) {
			t.Fatalf("installed offline setup missing %s before loading", required)
		}
	}
	for _, loading := range []string{"Preflight dependencies with", "sidefx acquire -out"} {
		position := strings.Index(text, loading)
		if position < start+end {
			t.Fatalf("offline setup is not before %s", loading)
		}
	}
	if !strings.Contains(text, "not a sandbox") {
		t.Fatal("missing loading qualification")
	}
	if runtime.GOOS == "windows" {
		t.Skip("shell execution of documented exports requires POSIX shell; installed ordering was checked")
	}
	// Execute the actual installed setup, not a harness-provided substitute.
	// No offline variables are inherited; approved tags/workspace must survive.
	dir := copyFixture(t)
	bin := t.TempDir()
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	goPath, _ = filepath.Abs(goPath)
	if err = os.Symlink(goPath, filepath.Join(bin, "go")); err != nil {
		t.Skip(err)
	}
	env := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "PATH", "HOME", "GOPROXY", "GONOPROXY", "GOPRIVATE", "GOSUMDB", "GOTOOLCHAIN", "GOTELEMETRY", "GOFLAGS", "GOWORK":
			continue
		}
		env = append(env, entry)
	}
	evidence := filepath.Join(t.TempDir(), "fresh")
	env = append(env, "PATH="+bin, "HOME="+home, "SIDEFX="+executable, "EVIDENCE="+evidence, "GOFLAGS=-tags=sidefx_offline_fixture", "GOWORK=off", "GOPRIVATE=private.invalid/dummy", "GONOPROXY=private.invalid/dummy")
	script := setup + "\n" + `test "$GOPRIVATE" = private.invalid/dummy` + "\n" + `test "$GONOPROXY" = none` + "\n" + `"$SIDEFX" acquire -out "$EVIDENCE" -direct -max-depth 0 -max-sites 100000 -max-edges 1 ./...` + "\n" + `exec "$SIDEFX" index -provenance "$EVIDENCE/acquisition.json" "$EVIDENCE/effects.json"`
	cmd := exec.Command("/bin/sh", "-ec", script)
	cmd.Dir = dir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("installed offline commands: %v\n%s", err, out)
	}
	record, err := os.ReadFile(filepath.Join(evidence, "acquisition.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, setting := range []string{`"GOPROXY":"off"`, `"GONOPROXY":"none"`, `"GOSUMDB":"off"`, `"GOTOOLCHAIN":"local"`, `"GOTELEMETRY":"off"`, `"GOFLAGS":"-tags=sidefx_offline_fixture -mod=readonly"`, `"GOWORK":"off"`} {
		if !strings.Contains(string(record), setting) {
			t.Fatal("documented setup did not preserve/record", setting)
		}
	}
}
