# sidefx session handoff

**Purpose:** Resume repository work and the function-first investigation program without repeating prior research, exposing unrelated work, or treating analyzer output as proof of a defect. This handoff records evidence and current state; it does not authorize fixes to third-party projects or start the next effectiveness experiment.

## 1. Project and rename state

Reffector is now named **sidefx**. The published repository is [`Version2beta/sidefx`](https://github.com/Version2beta/sidefx), repository ID `1422732752`; its default branch is `main`. Both `origin` remotes point to `https://github.com/Version2beta/sidefx.git`.

- `main` is at source-rename commit `45003edf09453a2ab3393f1c75ad1f73cc1b251b` (`Rename Reffector to sidefx`), pushed as a fast-forward and verified on GitHub. The commit changes source/product names, module/import paths, CLI, bundled skill/installer, man page, active docs, and archchar references; it preserves versioned `reffector.*` schema/provenance identifiers and historical evidence.
- Publishing `main` also published its five already-committed ancestors that were ahead of GitHub: packet-boundary tests (`9a046e4`), typed investigation persistence (`dea0d5f`), packet-reducer equivalence (`eee7d69`), Ralph tooling/campaign inputs (`5332b60`), and investigation documentation/checkpoint evidence (`3165846`). These were committed before the rename task; no uncommitted artifacts or `.ae` contents were included.
- The source-rename commit is separate from this handoff commit. The discovery branch's local source-rename work remains uncommitted and is not part of the published `main` commit.

Canonical local checkout paths after the worktree move:

| Path | Branch | State |
|---|---|---|
| `~/Development/sidefx` | `main` | `45003edf09453a2ab3393f1c75ad1f73cc1b251b`, pushed and tracking `origin/main` |
| `~/Development/sidefx-discovery-pivot` | `feat/effect-directed-discovery` | This handoff is the new branch tip; parent before this handoff was `95542cdc22d708a7c87dc604d7323ec364fdd837`. The completion response records the handoff commit SHA. |
| `/private/tmp/reffector-workflow-build-39a4830` | detached | `39a4830e4d04341ff920956c9260a61042097b35`; left untouched |
| `/private/tmp/reffector-workflow-milestone-39a4830` | detached | `39a4830e4d04341ff920956c9260a61042097b35`; left untouched |

Both development worktrees use the new `origin` URL. The feature branch is pushed only with its handoff commit; its outstanding source rename and other local work are not silently published.

### Local work deliberately left uncommitted

- In `~/Development/sidefx`, `.ae/journal.ndjson` remains modified and was not read or changed for this task. Untracked campaign artifacts remain, including `.runner.lock`, `DONE`, `logs/`, `run.log`, `run_state.json`, `reffector-ralph-8h.zip`, Python cache directories, and retained provider-event JSONL files under `reports/m5-*` and `reports/m8-*`. They were not staged or included in the source-rename commit.
- In `~/Development/sidefx-discovery-pivot`, the local source rename is still a working-tree change: old `cmd/reffector` and man paths are deleted while `cmd/sidefx` and `man/man1/sidefx.1` are untracked; `go.mod`, analyzer/output imports, README, `THIRD_PARTY_NOTICES.md`, and `docs/effect-directed-investigations.md` are modified. Existing docs edits are mixed with the local rename and were not committed here.
- The renamed function-first skill under `.agents/skills/sidefx-function-first/`, its helper/tests, `tools/`, seven Caddy/experiment reports under `reports/reffector-*.md`, and Python cache artifacts remain untracked in the discovery worktree. The helper test had one stale stats-key assertion corrected (`package_load_status` → `packages_selected`) and now passes; the skill itself remains local/uncommitted. No unrelated work was staged.
- The handoff commit contains only `docs/SESSION_HANDOFF.md`. Review `git status --short --branch` before any further staging. Do not inspect or modify `.ae` metadata without separate authority.

## 2. What sidefx analyzes—and what its output means

sidefx is a static Go source analyzer. It loads selected packages using the current Go build/workspace context, builds SSA, selects source functions in those packages, recognizes a finite vocabulary of direct operations and uncertain boundaries, relates functions with established and candidate call edges, then propagates summaries. Its default effect view is a bounded explanation of operation sites and reverse callers; it is not an exhaustive function census or a runtime trace.

### Function/effect semantics

- `effects[].direct: true` means the recognized category occurs directly in that function. `direct: false` means it is reachable through established calls with no direct occurrence. A function can have both a direct occurrence and an inherited path for the same category; the projection records that category once as direct. These flags are not exclusive path counts or ownership proofs.
- `possible_effects` comes from candidate relationships, not established calls. Keep it separate from established effects. `uncertain` is another field; neither false uncertainty nor an empty summary proves purity, safety, or absence of behavior.
- Function-view JSON does not include source locations or uncertainty reasons/provenance. Function identities may merge source declarations/variants. Use a focused location query or AST mapping before opening a selected body; do not guess locations from names.
- Established edges, candidate edges, unresolved interface/function-value dispatch, opaque callbacks, reflection and unsupported effect categories have different confidence. A candidate edge is a lead, not proof that a call occurs. Missing or truncated results are scoped to the queried packages, filters, build, and output bounds.
- `-max-depth`, site and edge budgets bound explanation/output, not package loading or all analysis cost. High bounds can consume substantial memory. Do not use large `-details`/graph exports by default.
- An empty or omitted function summary is not a purity claim. `-all` can show empty summaries; ordinary views omit some summaries. Tests are selected only with `-tests` and are not executed by analysis.

### Current CLI and names

The module is `github.com/Version2beta/sidefx`, command source is `cmd/sidefx`, the binary is `sidefx`, the man page is `man/man1/sidefx.1`, and the bundled standalone skill is `skills/sidefx/`. Common entry points:

```sh
sidefx -view functions -json -stats ./...
sidefx -view functions -all -json -stats ./...
sidefx -view effects -json -stats ./...
sidefx -unresolved -reason interface_dispatch -package "$pkg" -function "=$fn" ./...
```

Use `-view functions` to select a function prospectively, then focused source/contract/test evidence to decide whether behavior is defective. The serialized `reffector.*` schema/provenance keys and historical filenames/reports intentionally remain stable; they are compatibility/evidence identifiers, not a failure to rename the product. `tools/archchar` is a separate tool and its established protocol identifiers remain unchanged.

## 3. Function-first skill

The local project skill is `.agents/skills/sidefx-function-first/SKILL.md` in the discovery worktree. Its frontmatter name matches the directory, and its description identifies Go function-first defect investigation. The supporting `scripts/compact.py`, tests, and `references/validation.md` are part of the same standalone directory.

Completed improvements include:

- record a prospective function selection and question before reading its implementation;
- distinguish summary evidence, source inference, authoritative contract, runtime observation, and unresolved hypothesis;
- keep source-directed transitions explicit when the original concern is rejected;
- preserve established direct/inherited categories, candidate-only possible effects, uncertainty, and unknown location/provenance separately in the summary helper;
- bound and isolate authorized tests, record provenance/commands/results, and avoid equating a passing probe with a general safety claim;
- require a stopping record with elapsed/budget, investigated functions, open/parked leads, concrete stop reason, and supporting basis—without imposing exhaustive inventory or a finding quota.

**Validation:** five skill tests passed in 3.964 seconds using the current sidefx binary, including the optional real-CLI dry run and isolated author-known Go fixture. The fixture deliberately returns an error after overwriting a temporary file; it validates procedure and summary preservation, not blind discovery or real-world effectiveness. Retained evidence is under `/private/tmp/sidefx-skill-final.POEKJs/dry-run-m1es2g01/`. The test's first run failed because it expected an obsolete stats key; the assertion was updated to `packages_selected`, then the full suite passed. The offline Go suite in the discovery worktree also passed (`cmd/sidefx`, `internal/analyzer`, `internal/output`; `internal/effects` has no test files).

Pi documentation says project `.agents/skills/` are discovered from the working directory when the project is trusted, subject to settings, collisions, and host overrides. The skill validation note documents Pi 1.4.0 API/frontmatter checks; `pi --version` currently reports **1.5.0**. An actual fresh Pi startup, project-trust decision, startup skill listing, `/reload`, and interactive `/skill:sidefx-function-first` invocation were **not** verified. Do not claim Pi discovery yet. Before upgrading Pi further, open a fresh session from `~/Development/sidefx-discovery-pivot`, verify project trust and skill discovery/invocation on the installed 1.5.0 host, and check for name collisions. No Pi settings were changed and no upgrade was performed. The `sidefx` bundled skill is in the repository; its installation into a user skills directory was not performed or verified.

## 4. Evidence and investigation results

### Caddy, pinned commit `8e237f195506741d3109ba7792e81d57a5f65036`

Three different kinds of evidence must not be conflated:

1. **Coverage audit (static, not defect discovery):** all 2,630 selected source locations matched the analyzer inventory, represented by 2,629 emitted function identities because two closures share one identity. The supplied 2,223-function denominator was not independently reproduced: the audit found 2,208 named declarations and 2,683 functions including literals across all non-test source. Do not claim “2,223/2,223.” The current build omitted 47 declarations and six literals; 712 function summaries had no known effect/uncertainty and are not thereby pure. The default effect view truncated output; complete function coverage was a separate query. The audit measured very large detail output (about 1.9 GiB) and peak memory above 3 GiB, so avoid broad detail exports.
2. **Deep investigation:** confirmed a finite HTTP `GracePeriod` contract mismatch for ordinary `net/http` shutdown: after the deadline, active HTTP/1 responses were still able to complete although shutdown tracking ended. Repeated real local Caddy config reloads retained old-generation requests while the new generation served fresh requests. Separate tests reproduced a conditional Unix-listener ownership leak if an external pathname/permission change causes a post-bind `chmod` error; the mechanism is real, but natural trigger likelihood/deployment exposure is unknown. Neither finding was tied to a specific production incident.
3. **Function-first discovery:** after the original concern about `Upstream.fillDynamicHost` was rejected, a source-directed transition identified a conditional `MaxRequests` accounting defect. When a dynamic host's map entry expires after the one-hour idle interval while a request is still active, a later same-address lookup can create a fresh `Host` with zero active requests; an accelerated 20 ms/2 ms test reproduced that state split and availability past a limit of one. It did not run a real request held for an hour or establish production incidence. Two earlier function-selected leads were rejected.

**Material safety note:** an initial failed deep-investigation reload harness wrote Caddy autosave data at `/Users/rob/Library/Application Support/Caddy/autosave.json` and ran TLS storage maintenance in Caddy's default app-data directory. The prior contents and extent of changes are unknown. They were not read, restored, or removed; subsequent tests redirected those paths to temporary directories. Do not rerun the harness with defaults or overwrite/clean those locations. The deep report says to compare with a known-good backup or establish prior state before any action.

Reports in the discovery worktree: `reports/reffector-caddy-function-coverage.md`, `reports/reffector-caddy-deep-investigation.md`, `reports/reffector-caddy-function-first-discovery.md`, and `reports/reffector-caddy-complete-sites-depth1.md`. Detailed temporary evidence roots are recorded in those reports, including `/tmp/reffector-caddy-coverage-mwyoxO` and `/tmp/reffector-caddy-function-first.jfJeE9`.

### DRPC

Independent verification used byte-identical copies of the relevant production files and isolated tests; the original 134-file source manifest was unchanged. Results:

- **Confirmed:** `drpcpool.Pool.Put` same-key replacement at `Capacity:1`, `KeyCapacity:0` can remove the key mapping during eviction and append the replacement only to global order. `Take` then fails despite global count 1; a later Put panics through the orphaned state. This violates the documented Put/Take/cache contract.
- **Reproduced, contract uncertain:** a stale expiry callback can remove a replacement after a controlled `Close`/Put interleaving; the package docs do not define the timer-identity guarantee precisely enough to call this a confirmed contract violation.
- **Confirmed:** a `drpcctx.Tracker` callback that calls `runtime.Goexit` skips the non-deferred `WaitGroup.Done`, leaving `Tracker.Wait` blocked despite the README promise that it waits until callbacks exit. No Server-level shutdown reproduction was run; that consequence remains source-inferred.

The original package's offline dependency attempt was blocked by missing cached modules; no network fetch occurred. Isolated copied-source tests passed or failed at the intended contract assertions. Full evidence: `/private/tmp/drpc-verify-20261102-NKJjjM/verification-report.md`.

### Gorilla WebSocket v1.5.3 paired trial

The source snapshot was byte-verified against the v1.5.3 module. Independent replays established three strict contract defects across the pair: cancel-only context does not interrupt a stalled post-connect handshake; the client accepts a subprotocol it did not offer; and compression activates from an unsolicited extension response. Cancellation was found by both arms; the other two were each found by one arm. The compressed `SetReadLimit` bypass was reproduced by both but remains contract-qualified because the docs do not specify compressed-versus-decompressed accounting. Raw response-header-key injection was reproduced, but attacker control of the application-supplied map was not established; non-minimal frame length acceptance was reproduced without a demonstrated consequential receiver obligation.

The treatment and control each had two arm-level strict findings, sharing cancellation, for three unique strict defects. The different unique findings were adjacent handshake checks that source-only review could plausibly find. This one pair demonstrates **no incremental confirmed-finding advantage or causal effectiveness claim** for sidefx. The treatment used more turns/time and recorded usage; those differences are descriptive, not attributable to sidefx. The retrospective session-boundary times are 11m51.718s versus 9m55.229s; Pi metadata records 95 versus 77 turns and `$0.09216014` versus `$0.07385738`. These recorded costs are not verified provider charges. An earlier report gives shorter manual elapsed times; use the retrospective timestamps and preserve that discrepancy. Full reports/evaluations: `/private/tmp/reffector-effectiveness-pgx.QwnFoM/`.

### What these results justify

The Caddy and DRPC behavior claims have focused source/contract evidence and the reproductions described above; they do not establish broad production incidence. The Gorilla pair supports only its bounded, independently replayed outcomes, not a general sidefx effectiveness conclusion. The synthetic skill dry run demonstrates procedure/format behavior only. None of these reports establishes novelty unless explicitly stated (none did); novelty/history checks were not performed.

## 5. Validation and current publication boundary

The source-rename working tree passed `go test ./...`, `go vet ./...`, CLI build/help/function-view smoke checks, all 40 archchar tests, the effect-investigation packet-boundary test, nested Go helper checks, and `git diff --check`. The separate discovery worktree's offline `go test ./...` also exited 0. The skill's five tests, including its isolated real-CLI fixture, passed after the stats assertion correction. GitHub verification confirmed the renamed repository identity and `main` at `45003edf...`; both remotes were fetched and local `main` now tracks `origin/main`.

The published rename commit does not include the discovery worktree's uncommitted mirror of the rename, the function-first skill, its reports, or campaign artifacts. The handoff commit is separate and contains only this document. No third-party source was modified. No new controlled effectiveness trial was run.

## 6. Next controlled effectiveness experiment (planned, not started)

Use a fresh, previously unexposed target snapshot; do not reuse Caddy, DRPC, or Gorilla findings as blind-discovery targets. Before either arm starts, pin and hash the source/build context, scope, contracts available to both arms, exact model/provider/version/reasoning, prompts, tool permissions, wall-time/turn/token budgets, and stopping rule. Randomize arm order; keep sessions independent and share no candidate hints or outputs. The only intended difference is sidefx inventory access in treatment versus ordinary source navigation in control. Record prospective selections before implementation reads, source-directed transitions separately, every command and exit status, time/usage evidence, and the skill's stop record.

Have an evaluator blind to arm identity adjudicate deduplicated candidates against authoritative contracts and independent bounded probes. Report strict confirmed defects separately from reproduced/contract-uncertain behavior, exploitability/impact-uncertain leads, and rejected hypotheses. Compare validated unique findings and time/usage descriptively; a single pair is not a general causal result. Pre-register what counts as a finding and how unresolved/low-severity leads are handled. Do not begin until the scope and all isolation/side-effect paths are checked and authorized.
