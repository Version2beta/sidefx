# Manual release preparation

There is no automated binary publication workflow. `.github/workflows/ci.yml`
builds, tests and vets on Linux, macOS and Windows using `go.mod`'s Go version;
verify actual candidate results rather than treating configuration as a pass.

## Approve the candidate first

- Review the complete candidate file list, including embedded `skills/sidefx/SKILL.md`
  and the three references embedded by `docs/embed.go`. Include their canonical
  sources. Do not distribute private logs, sessions, copied targets, caches,
  archives or historical investigation evidence merely because they are present
  or tracked. Curating those files needs owner review; do not rewrite the records.
- Approve a version and supported Go/OS/architecture claims. `go.mod` requires
  Go 1.26.0; a newer local toolchain or a successful cross-build does not validate
  the minimum toolchain or another platform's runtime behavior.
- Review the deliberate CLI compatibility changes in README: the default is the
  effect view; use `-view functions -json` for function arrays. `index` and
  `acquire` are reserved first commands. Native indexing requires effect-view
  schema 1, not function arrays or audit JSONL.
- Commit an owner-approved candidate before tagging or publishing. A dirty-tree
  validation binary is not an immutable release, even when its tests pass.

## Validate and build without installing

From the approved checkout, choose an existing external output directory and
set `OUT` to its absolute path. Use approved dependency/build settings. For an
already cached, standalone module, these offline settings prohibit Go dependency
fetching; missing dependencies/toolchains are blockers, not permission to repair:

```sh
export GOPROXY=off GONOPROXY=none GOSUMDB=off GOTOOLCHAIN=local
export GOENV=off GOWORK=off GOTELEMETRY=off GOFLAGS=-mod=readonly
# GOWORK=off is for an approved standalone-module scope only.
go test ./... -count=1
go vet ./...
go build -trimpath -o "$OUT/sidefx" ./cmd/sidefx
"$OUT/sidefx" -help
go version -m "$OUT/sidefx"
```

Use `sidefx.exe` on Windows. The shell examples above use POSIX syntax; adapt
variable assignment and paths for the chosen shell. No custom version flag or
linker version variable is defined: retain `go version -m` output, exact source
revision/dirty identities, toolchain/build settings and the executable SHA-256.
A VCS revision alone does not identify uncommitted content. Repeat the build and
compare hashes under identical inputs if claiming local byte reproducibility;
this does not establish reproducibility across toolchains or hosts.

### Eventual alpha metadata (owner-authorized only)

Use Go's supported tag-derived module build metadata rather than adding an
application version flag or injecting a proposed release name. For the root
module, the eventual tag is `v0.1.0-alpha.1` (not `cmd/sidefx/v0.1.0-alpha.1`).
Only after the owner approves a committed, validated revision, create that real
tag on that exact revision and build from its clean checkout using the command
above, with output outside the checkout. Verify `go version -m` reports module
`github.com/Version2beta/sidefx` at `v0.1.0-alpha.1`, the approved `vcs.revision`,
and `vcs.modified=false`. Retain fresh executable and archive checksums for
approval; earlier untagged binaries are not interchangeable with that build.
Creating/pushing the commit or tag, running remote CI and publishing each require
separate owner authorization; these instructions do not perform those actions.

The VCS-aware local build mechanism is verified with Go 1.27.2; check actual
metadata with the selected release toolchain, including Go 1.26 if used. An
untagged checkout can report a pseudo-version; uncommitted edits add `+dirty`
and `vcs.modified=true`. A source extraction without Git metadata can report
`(devel)`. None of these is an assigned alpha version. Application help is not a
version interface, and an injected label would not establish a real module tag
or clean source provenance.

## Validate the distributable, then obtain publication approval

- Stage the **new** executable in a private temporary directory, with isolated
  HOME, and run it from outside the checkout. Check a small known-source fixture,
  JSON/function views, errors, fresh `acquire -out NEW_DIR -direct ./...`, saved
  `index`, exact category/site retrieval and complete cursor pagination.
- With Go and Python absent from PATH, saved indexing and `skill install` should
  still work. Fresh analysis needs Go and the selected build's dependencies.
  Skill installation does not install the executable. Verify every installed
  reference/link and refusal to overwrite modified/unmanaged files or symlinks.
- Run the full Go suite, vet and relevant regression tests on the candidate with
  controlled time/memory and offline settings. Check actual minimum-toolchain
  and supported-platform runs. Do not run old experiment/target runners as release
  checks. Output caps do not bound loading cost; empty results are not purity.
- For each approved binary package, include `LICENSE`, `THIRD_PARTY_NOTICES.md`,
  README and `man/man1/sidefx.1` as appropriate, and retain the embedded skill
  resources. Generate SHA-256 checksums for the actual files shipped, record the
  exact build command and test evidence, and verify an extracted package outside
  the checkout. No automatic PATH or man-page installation is provided.
- The owner approves final source curation, version, platform claims, commit,
  tag, archive/checksum review and publication. Publishing and a real
  `go install ...@VERSION` validation are separate authorized online actions;
  offline local builds do not prove module availability at `@latest`.
