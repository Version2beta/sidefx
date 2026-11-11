# Operation recognition decisions

Baseline `9526abb25e97f0c065d1565b06b696debf49781a`. Persisted before source edits. The [complete audit matrix](operation-coverage-audit.md) retains O01–O27/E01–E32 and all third-party/native investigation scope. The finite core below is **implemented** and tested; optional/deferred siblings are listed separately. Matrix baseline labels do not retroactively describe additions as pre-existing coverage.

## Implemented deterministic anchors

| Decision | Baseline → implemented change | Contract and evidence |
|---|---|---|
| D01 Root flags | Baseline unconditional union → exact call-arity constant/dynamic flag handling in LookupCall; original receiver-signature offset rejected after P1 review | 3 source args when bound, receiver + 3 for direct/expression; malformed arity union+Unknown. Preserve unknown-flag union/resource acquisition, not permissions proof. Bound write/dynamic/mode-0601, expression, inheritance/candidate/filter/source regressions added red before repair. |
| D02 fmt opacity | Baseline skipped variadic args/unconditional Append* zero → exact layouts and bounded literalArgumentValues opacity proof | Correct exact argument layouts, conservative literal-array proofs; basic memory formatting zero, opaque formatter Unknown, no callback execution edges |
| D03 alternate FS | Add File.ReadFrom/WriteTo, CopyFS, filepath traversal, public io/fs facades/contracts, zip open/close | File.ReadFrom writes receiver; WriteTo reads receiver; peers generic/opaque; virtual FS need not OS-backed; ZIP open not extraction; public declaring identities preferred |
| D04 stream IO | Add io.read/io.write contracts/helpers, finite bufio, JSON/XML/Gob streams, template execution, gzip/flate/zlib/tar, fmt and builtin printing | Stream attempts may be memory-only, no fs/net destination inference; constructors/memory Marshal/Unmarshal/Append not IO; callbacks remain opaque; codec close not underlying release |
| D05 HTTP | Helpers response, header preparation, route registration, ReverseProxy | Preserve ServeFile fs.read; Header mutation is not on-wire output; register not invoke/listen; constructors negative |
| D06 logging | log/Logger/slog attempts; Fatal retains exit; Panic also control.panic | Submission not durable audit; Writer/Handler/format/value/state opacity; no application-name matching |
| D07 network/RPC | Socket declaring contracts/data methods, DNS lookups, RPC dial/client operations | net.Conn can be custom/memory-backed; lookup not connection; RPC Call/Go generic IO over codec, no invented new connection or service effect |
| D08 TLS/random | Key-file reads, handshake attempts, crypto/rand.Read/Text | Completed handshakes may no-op; verification hooks/state and mutable crypto/rand.Reader override remain opaque; no blanket pure crypto verification/signing/parsing summary or adequacy inference |
| D09 try-lock | Exact Mutex/RWMutex TryLock/TryRLock | Lock attempt plus state_condition, never successful acquisition/ownership inference |
| D10 time/deadline | Now/Since/Until; net/os/tls/ResponseController deadline attempts | Ambient clock vs pure Time arithmetic; deadline may fail/clear/forward. WithDeadline/WithTimeout and Cause variants report possible scheduling plus state_condition (not candidate effects); nil parent excludes scheduling, nonpositive timeout retains parent opacity without timer. WithCancel/Background/TODO not timer operations |
| D11 load/credentials | plugin.Open and finite syscall credential setters | Never execute fixtures; state/platform/init opacity; no inferred initializer effects, successful drop, or uniform thread/process scope |
| D12 sync | File.Sync adds fs.sync, keeps fs.write | Flush not fsync; sync attempt not durability/crash-order proof |
| D21 panic | Exact ssa.Panic and log.Panic | control.panic not process.exit; recover/defer context not successful containment/cleanup |

Implementation uses exact resolved package/receiver identities and existing SSA/sites/propagation. No policy/taint/alias/escape analysis, universal call inventory, callback execution inference, vulnerability labels, or schema/propagation redesign. Approved new names: `io.read`, `io.write`, `log.write`, `net.read`, `net.write`, `net.lookup`, `tls.handshake`, `crypto.random`, `fs.sync`, `io.deadline`, `clock.read`, `process.load`, `process.identity`, `http.handler.register`, `http.header.write`, `control.panic`. Append only: baseline 48 bit positions including Unknown → all 64 used. `internal/effects/effects_test.go:TestSetCapacityAndUniqueNames` guards capacity, one-bit membership, unique labels and hidden Unknown; further expansion needs a deliberate representation change.

## Current contracts retained

- D13 callbacks: AfterFunc register/stop, Once invocation barrier vs constructor, WaitGroup.Go launch/accounting, Map.Range/Pool.Get/CertPool hooks keep opacity. Registration is not execution. Finite function alternatives and VTA remain possible-only; open external callers stay open.
- D16 semantic/pure context: authorization/tenant/secrets/injection/CSRF/TLS assurance/races/deadlock/leases/retry/cleanup/errors/durability/cardinality/compatibility/containment/fallback remain reasoning obligations. Ordinary arithmetic/path/key/config/parse/compare/time conversion is context, not automatically effect. Pure-looking APIs may call hooks/environment: do not blanket suppress Unknown.
- D17 output: normal view for classified effects; unresolved exact API for ID-only calls. A-zero/absent instructions may have no site. Preserve legacy function-pair text omissions; recommend JSON for possible/unresolved evidence. `candidate_function_value` and `candidate_vta` are distinct from established calls. Runtime conditions are not candidate possible_effects.

## Explicit deferred gaps (not safety or blanket C)

| Decision/gap | Reason / retained evidence boundary |
|---|---|
| D03 optional syscall file subset | Open/Read/Write/Close/Fsync/Rename/Unlink/Chmod/Chown are worthwhile B but not added here: platform/FD/flag semantics need a separately bounded contract/fixture selection; no arbitrary syscall/cgo promise. File.Sync is implemented independently. |
| D04 finite stream subset | Other exact stream APIs, e.g. fmt.Scan*, remain unclassified; selected fmt coverage is printing, not a blanket package summary. These can be worthwhile future B anchors, not safety or semantic C. |
| D07 RPC server extras | Register/serve/accept/codec APIs can be exact anchors; client subset selected first. No implicit arbitrary service-name semantics. |
| D11 other globals/native | Logger/runtime setters, namespace/capabilities/native dlopen need finite exact API/platform/module evidence; arbitrary setter/global-store names are not operations. |
| D14 native map/store/allocation projection | SSA structural anchors practical, but classifying every local map/store/allocation as concurrency/external effect would mislead and flood output. Separate structural product decision; sharedness/alias/correctness C. |
| D15 external adapters | Cache/object/NoSQL/LDAP/ORM/session/token/lease/broker/payment/jobs/telemetry/native alloc/free exact APIs can be worthwhile B, not all C. No verified finite module/version catalog or downloads authorized. Source bodies/open contracts still available. |
| D18 broader stdlib VTA | Default discovery bounded to local receiver/source universe; general stdlib candidate summary path limited. New table entries do not guarantee field/return interface coverage. Expansion needs budget/performance evidence; no fabricated edges. |
| D19 argument roles | Optional fixed positional enrichment practical, but independent from operation sites; mutable Cmd/HTTP/heap/capture fields remain unresolved. No general taint/alias inference. |
| D20 SQL ownership/taxonomy | Exact Begin/Prepare/Commit/Rollback already retrievable under db.exec. Rows/Stmt/Tx/HTTP body ownership nuances need specific return/error contracts; no SQL parser/isolation model or new categories selected. |
| Private implementation receivers | Public contracts preferred; private OS/net receiver support only where specifically needed and verified. No private-name heuristics or full stdlib traversal. |

## Test and guidance contract

Static-only exact stdlib fixtures cover lookalikes, pure constructors, memory-backed streams, conditional controls, wrapper/reverse callers, open interfaces, finite function alternatives, exact API/effect/site/operation/reason filters, JSON/text and exact call locations/IDs. No plugin/process/credential/network operation function is executed. Generic Writers now have io.write while retaining the no-inferred-fs/net/target guarantee. Focused tests, bounded full offline Go suite/vet/build and CLI-to-source smoke were run; results and initial compatibility failures are below.

Canonical embedded guide: `skills/sidefx/SKILL.md` via `embed.go`; installer copies to managed `~/.agents/skills/sidefx/`, never edit home/installed copies. `.agents/skills/sidefx-function-first/` is independently maintained. README/man/canonical and function-first skills now describe actual vocabulary, ID-only vs classified retrieval, possible-only text omissions, function-value candidates, registration/header/stream abstractions and semantic limits. They link this full catalog without imposing exhaustive sweeps or a new workflow. Installed/generated/home copies were not edited.

## Actual implementation / validation

- `effects.go:LookupCall/LookupInterface/literalArgumentValues`, `operations.go:addOperationSummaries`, `analyzer.go:ssaOperation` implement the selected exact anchors; `uncertainty.go:summaryReason` retains interface/state/callback distinctions. No schema or propagation redesign. Lookup table membership still does not promise unsupported candidate discovery.
- Format proof rejects aliases/competing/repeated uses, late/multiple/unproven initialization, nonconstant bounds and named basic types with possible formatting methods. Proof storage is bounded by actual SSA references before allocation, including a huge-source-array regression control. No callback effects or edges are invented.
- Context constructor timer summaries fit existing state_condition+Unknown semantics: parent deadline/cancellation/time/custom hooks remain opaque, not successful scheduling or callback execution. Cancellation-result provenance is unchanged. Handshake/plugin summaries retain callback opacity; state/platform/no-op limitations are contract caveats, not inferred state facts.
- Exact finite API lists are source-maintained in addOperationSummaries/LookupInterface; the shared audit table maps them to all 59 families. Tests: `cmd/sidefx/operation_coverage_test.go` (three TestOperationCoverage* tests), `context_timer_operations_test.go`, `credential_operations_test.go`, existing compatibility tests, and `internal/effects/effects_test.go`.
- Initial implementation, before independent review, on offline Go 1.27.2 darwin/arm64: focused rounds passed; final `go test ./... -count=1 -timeout=300s` passed (CLI 96.087s), `go vet ./...` and build passed. Linux/Windows **source loading** tested credential availability; no foreign binaries or dangerous functions executed. Four function-first helper checks passed, optional probe/defect dry run skipped. Fifteen final CLI-to-source smoke invocations passed. Full suite's first run failed seven stale expected-coverage assertions; corrected to the newly modeled operations, preserving no-fs/net mislabel and no-invented-target guarantees. No unchanged deterministic failure was simply rerun.
- Evidence: `/tmp/sidefx-operations-implementation.QgMteL` (full/focused/build/vet logs and exit codes); `acceptance-smoke/` has commands, environment, hashes, copied fixture and source-location/candidate/filter outputs. Pre-review binary SHA-256 `a188937a695c2963d0307cfd71766f80725a7f2045955709e06df3f8b3ec0037`; superseded for D01 by the correction below.
- Post-review D01: independent read-only review found that bound Root.OpenFile arguments exclude the receiver although types.Func retains it. The old bound readonly/mode-zero test masked permission-as-flags misclassification. `TestRootOpenFileBoundArgumentShapes` and `TestOpenFileMalformedArityIsUncertain` failed before production repair, then passed after exact 3/4-argument handling; no receiver/alias inference. Supported-shape dynamic flags retain the existing conservative union, while unsupported arities add Unknown. Ticker.Reset already reads its final argument and needs no related repair. New evidence `/tmp/sidefx-bound-correction.p8MgpQ`: focused red/green, full offline suite (CLI 93.906s), vet/build, 15-query static smoke and all three independent failing probes corrected. Current binary SHA-256 `69fd1ad58a9a78435574aaa16406e46d15694d86ccf3cbed6ea3ae6ab74d2a67`; current hashes/commands in that directory supersede earlier manifests. Independent verification remains pending, not acceptance.
- Compatibility: existing IDs/schema/bit assignments and legacy pair-text format preserved. Newly classified calls intentionally move out of unresolved-only views and increase effect/site counts; File.Sync retains fs.write. All 64 bits are occupied. Output filters/budgets still do not cap package loading; local formatting proofs inspect source-local SSA references/block order. No project-size performance or exhaustive Go1.26+/GOOS claim.

## Subsequent category-index/workflow implementation

The approved [complete index and investigation workflow](operation-index.md) adds
native installed `sidefx acquire` and `sidefx index` over saved effect-view JSON,
plus a self-contained optional skill/reference package. The older Python helper
is development/reference-only; the normal installed workflow requires no Python
or source checkout.
It changes no recognition vocabulary, SSA/propagation, IDs, analyzer schema or normal CLI
limits. `-all` remains function inventory, not unlimited origins. Every emitted
site category is indexed once without sampling; complete declared-query orientation
additionally needs recorded unfiltered acquisition provenance and verified no site
loss. Legacy/filter/capped/unknown inputs are labeled honestly. Exact-category/ID
pagination preserves original effect/evidence and caller uncertainty. The
canonical and function-first skills support focused-question routes that skip
complete category orientation while recording survey limitations, and broad
prospective-discovery routes that require complete orientation before selection.
Both preserve bounded behavioral follow-through and local contract verification;
installed/home copies remain untouched. This supersedes only the
prior statement that no new workflow was imposed, not the historical recognition
validation or any of its limits. No detection-effectiveness claim follows.
