---
name: sidefx
description: Retrieve and interpret Go effect sites, reverse callers, and unresolved boundaries with sidefx. Use for effect investigation, architecture, change impact, test planning, or code review; effects are not defect findings.
metadata:
  version: "installed-operation-index-2"
---

# sidefx — investigate operations

sidefx reports recognized possible operations and static reachability, not
execution or defects. Use evidence to answer the user's question; do not impose
a refactoring workflow. Shared categories do not justify consolidation, and
separate implementations may be appropriate. Investigation does not authorize
source edits, dependency repair, downloads, or running the analyzed application.

## Scope and loading

Use a verified installed binary. Check `command -v sidefx`; if missing, stop and
obtain the authorized executable. The skill does not install the executable.
Add its directory to PATH or replace **every** `sidefx` invocation with its verified
absolute path, for example `/opt/sidefx/bin/sidefx index REPORT.json`. Do not assume
a checkout, Python or automatic dependency installation. Saved indexing and skill
installation need only the binary; fresh analysis also requires Go 1.26+ on PATH.

Check `sidefx -help`; a preinstalled binary may differ. Run analysis from the
target Go module/workspace with an explicit, narrow package scope. `./...`
selects recursive subpackages; local directories also recurse. `-exact .`
selects only the root package, but does not override an explicit `./...` pattern.
Respect existing workspaces, replacements and build settings; `-tests` includes
test-file functions without executing tests.

Before any preflight or fresh loading, use the approved offline environment;
preserve existing tags, replacements and workspace settings:

```sh
export GOPROXY=off GONOPROXY=none GOSUMDB=off GOTOOLCHAIN=local GOTELEMETRY=off
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly"
# Preserve GOPRIVATE, tags and GOWORK; disable only the private download bypass.
# Set GOWORK=off only for an approved standalone-module scope.
```

Loading is not a sandbox: Go tooling can write caches and raw workspace commands
can update sums. GONOPROXY=none prevents inherited GOPRIVATE/GONOPROXY patterns
from bypassing GOPROXY=off. This is the approved Go download policy, not a network
sandbox or arbitrary target execution permission. Missing dependencies/toolchains are blockers, not
authority to download, repair dependencies or silently retry with another policy.

Preflight dependencies with `go list -mod=readonly -deps ./...`, using the same
scope/environment (add `-test` for `-tests`). In a workspace prefer sidefx's
own loading as the preflight: raw Go tooling can write `go.work.sum`, whereas
sidefx uses temporary workspace metadata. The settings above keep Go and
sidefx loading offline; missing cached dependencies/toolchains are blockers.
Dependency listing is not type checking: sidefx must also load successfully.
Loading errors invalidate the requested analysis, not establish zero effects;
report the blocker rather than silently narrowing scope or repairing modules.

## A — Complete category orientation before selection

Establish authority, target/dirty-source and producer identities, package/build
scope and query/source-reading/reporting budgets first. Prefer qualified saved
effect-view JSON; keep full records outside context. The installed CLI alone does
complete native indexing, without a skill, Python, checkout, Go or additional SSA:

```sh
sidefx index -provenance "$EVIDENCE/acquisition.json" "$EVIDENCE/effects.json"
```

For an authorized fresh acquisition from the target cwd, set EVIDENCE to a **new
nonexistent run-directory** under an existing authorized parent. Choose direct
selected roots or reachable dependency/candidate origins explicitly; this example
chooses direct roots. The directory and every output are created exclusively;
rerunning with an existing path fails before analysis or writes.

```sh
sidefx acquire -out "$EVIDENCE" -direct -max-depth 0 -max-sites 100000 -max-edges 1 ./...
```

Read the installed [acquisition/index contract](references/operation-index.md).
It records command/cwd/binary/build/report identities and saves effects.json,
effects.stderr, acquisition-start.json, acquisition.json and index.json. Retain
source revision/dirty identities separately. Check successful loading, equal
matching/emitted counts and false site truncation; capped output returns nonzero
with an incomplete index, not complete orientation. No silent retries/scope
changes. `-all` adds function inventory, not unlimited sites/edges. Direct scope
excludes dependency origins/candidate-only helpers; reachable scope still does
not cover all dependency bodies/builds. Caller depth/edge loss is separate from
category loss. `acquire` is unfiltered recognized-operation acquisition; normal
CLI retains filters/contracts/unresolved queries, and recognized reports do not
contain a complete unresolved inventory.

Read **EVERY exact category ID plus coverage/bounds in the index**, not all sites.
The native index includes all labels present in emitted established or possible site
fields once, including singletons, without samples or category/frequency caps.
Counts are distinct canonical site IDs per category; multi-effect membership
**overlaps** across categories. Origin selected/dependency/unknown counts come
from matching reverse origin nodes, not direct/inherited function marks.
Established origin labels do not establish a particular candidate caller path.
Uncertainty overlaps both established/possible labels. Truly uncategorized
unresolved sites are separate; unresolved-view records can also have possible-only
labels. Identified unresolved counts are not an emitted inventory. The registered
vocabulary/recognized-zero categories are unavailable from schema 1, not zero.

Without qualified provenance, `declared_query_categories` is `not_established`:
legacy/filter/capped JSON cannot silently become complete selected-scope coverage.
`emitted_records=complete` only means the projection lost none of its valid input.
Retain analysis budgets, scope exclusions and graph depth/edge/site bounds.
The index is not a risk score, absence proof or exhaustive mutation inventory.

## B — Select a small prospective portfolio, then retrieve

Across the whole index choose a few consequential, locally tractable obligations,
not the largest counts or familiar categories. Keep one reasoned alternative lead.
Consult relevant portions of the installed identity-qualified full [59-family
coverage reference](references/operation-coverage-audit.md) and [recognition decisions](references/operation-recognition-decisions.md); do not regenerate it or sweep it
as a checklist. If required authority/reference inputs are absent, stop for them.
Before source reading record site/API/category/location and uncertainty, why it
matters, provisional obligation, falsifier, strongest protective explanation,
minimum next evidence and stopping bound. No finding quota.

Retrieve an **exact** category or canonical ID in small deterministic pages:

```sh
sidefx index -category fs.write -limit 8 "$EVIDENCE/effects.json"
sidefx index -category fs.write -limit 8 -cursor CURSOR "$EVIDENCE/effects.json"
sidefx index -site SITE_ID "$EVIDENCE/effects.json"
```

Copy `next_cursor` unchanged with the same selector until null when a full set is
needed. It binds report SHA-256, category/site selector and offset; stale input or
wrong-selector continuation fails. Alternatively follow `next_offset` with
`-offset` and `-expect-sha256` using `input_sha256`. No first-24 cutoff. `total`
means corresponding sites in that saved input. Pages retain original sites, APIs,
operations/evidence, reverse nodes/edges/contracts and caller uncertainty. Saved
retrieval reparses JSON without SSA. Site count is bounded (1–100/page), not graph
bytes; graphs retain original limits. Missing caller context needs an authorized
focused query, which reloads analysis.

### Focused CLI queries

Adapt the explicit scope to the question; these examples use the target module's
`./...`. Choose a query rather than treating these as mandatory steps.

```sh
# Locate filesystem-write sites without expanding callers.
sidefx -effect fs.write -max-depth 0 -max-sites 20 ./...

# Examine an exact response-output API and its reverse callers.
sidefx -api 'net/http.(ResponseWriter).WriteHeader' -max-depth 2 -max-sites 20 -max-edges 50 -json ./...

# Explore callers of unresolved interface boundaries, excluding other reasons.
sidefx -unresolved -reason interface_dispatch -max-depth 2 -max-sites 20 -max-edges 50 ./...

# Add available method-set implementations as investigation leads.
sidefx -contracts -unresolved -reason interface_dispatch -operation '.Save' -max-depth 2 -max-sites 20 -max-edges 50 -json ./...
```

Start with consequential categories → direct operation sites → source locations
and reverse callers. `-effect` matches category substrings (`fs.rename` and
`resource.release` are distinct from `fs.write`); `-api` is exact, `-operation`
is API-substring search. Copy a canonical ID into `-site ID` to expand one site;
keep `-unresolved` for an unresolved-only site. IDs survive filtering, not source
location changes. Multiple callers can share one direct site; independent calls
to the same API remain separate sites. Multi-category sites are not duplicates.

Depth 0 shows origins; 1 adds immediate callers. `-max-sites` and `-max-edges`
bound output, not loading cost. Check report and per-site `truncated` flags;
cycles/revisits are references, not omissions. Graphs show supported reachability,
not all execution paths. Package/function filters restrict roots and reachable
callers; remove them for broader caller context. Prefix a full local function
name with `=` for one exact root, excluding uncalled closures as separate roots.
`-direct` keeps only operations in the selected roots, not helpers or candidate-only
effects; in function view it emits recognized direct effects only, while
uncertainty still propagates. Explicit unresolved leads, unfiltered inventory
and contract-implementation summaries are not suppressed by `-direct`.

## Interpret and verify

Default text and `-json` describe the same effect view. JSON groups reference
sites; reverse graphs reference nodes and edges. Follow IDs, not same-spelled symbols
across generic/test variants. Sites carry API identities, locations, recognition
and uncertainty; `operations` preserves API-specific evidence. A declaration
boundary is not a source call. Inherited reachability is not a new direct site.
Reverse edges have `resolution: established` when SSA directly supports the
call, or `candidate_vta` when Go VTA follows values through assignments,
arguments, returns, fields, or callbacks. `candidate_function_value` instead
identifies finite alternatives from bounded function-value provenance. A
candidate is not proof of execution; open boundaries remain open. Function JSON
keeps established effects in `effects` and candidate-only paths in
`possible_effects`, even when `uncertain:false`. Compatible function-pair text
omits possible-only and uncertainty-only rows; use JSON to preserve them.
Function-view `-effect` matches and filters both lists, preserving their labels.

Uncertainty is separate from effects. `interface_dispatch` means the concrete
implementation remains unresolved: a named contract summary (such as response
output or `io.Closer` release) supplies no concrete target or resource destination.
`-unresolved` selects only effect-free unresolved leads; known-effect uncertain
sites remain in normal groups. `-reason` filters existing reasons in either
selection; it does not classify significance. Remove filters to inspect other
origins, including opaque calls and callbacks. `-contracts` adds available
`contract_compatible` method-set implementations beside `established` and
`candidate_vta` targets. Contract-compatible leads are not call edges and the
available type universe may omit external implementations. Implementation
`effects` use established calls; `possible_effects` (text: `possible=`) require
candidate edges. Source-site contracts aggregate the strongest relationship and sorted
effect unions across generic/test variants, excluding established effects from
the possible-only list. This does not merge graph nodes or close open boundaries.
Never invent implementation edges.

Check coverage: selected packages/functions, selected versus dependency sites,
identified versus matching versus emitted counts. `-all -json` adds unfiltered
source inventory; `-stats` also writes coverage to **stderr**. Empty matches,
omissions and `uncertain: false` do not prove absence or purity. Analysis covers
the current build and a limited API vocabulary, not arbitrary dataflow or mutation.
Known identity is not classified effect: try `-unresolved -api EXACT` for an
unclassified API; normal `-api` queries can be empty. A-zero boundaries and
unmodeled native instructions may have no site at all. Summary support does not
promise arbitrary stdlib interface-field/return or candidate propagation.

`io.read`/`io.write` are stream attempts, possibly memory-backed, not fs/net
destinations. Header preparation and handler registration are not on-wire output
or handler invocation. Logging is attempted submission, not durable audit;
Flush is not fsync; panic is not necessarily exit. Runtime/state-conditioned
operations carry uncertainty, not candidate `possible_effects`. Read source for
policy, identity, lifecycle, ordering, errors and adequacy; no category answers
those obligations. The [local source coverage matrix](references/operation-coverage-audit.md)
and decisions retain full family scope and explicit deferred gaps; consult the
relevant part as needed, not a mandatory catalog sweep.

## C — Follow a bounded behavioral connection

Read the operation's controlling modes/decisions and nearest relevant callers or
callees. Follow the most consequential unanswered local connection, not every
question: persistence completion **and retirement of recovery sources**; other
accesses to the **same object** and caller-lock assumptions; borrowed retention,
copy/ownership and producer reuse (first check retention is enabled); caller
return/skip/retry/fallback after errors; queued remainder, consumed/committed and
restart positions **after rejection**, including one next iteration; cleanup and
recovery representations and which evidence survives reconstruction. Creating a
closure is not invoking it. Validate candidate dispatch/configuration before
claiming execution. No architecture sweep, all-lock/error analysis or separate
vulnerability brainstorm. Park a lead if the next boundary exceeds its budget.

## D — Local contract, counterevidence and permitted verification

Read the nearest interface/comments, relevant test expectations and protective
branches. Separate authoritative guarantees, inferred expectations, source
observations and runtime results. Seek mode gates, recovery and intentional
tradeoffs; test source is not a run. Only separately permitted focused checks
may execute. Distinguish **confirmed contract violation**, **reproduced behavior,
contract uncertain**, **source-supported hypothesis**, **rejected hypothesis** and
intentional behavior. Missing contract/safe verification is a productive stop,
not permission to invent guarantees. Honor provider/safety blockers without
retry, rephrasing or alternate routes. Consider the alternative only if authority
and budget remain; reserve reporting time.

## E — Exposure and stopping record

Log identified operations versus index categories actually presented, sites
retrieved, leads selected before reading, source ranges, actual subquestions,
contracts/counterevidence, commands/statuses/tests and productive stops. Record
binary/source revision and dirty identities, cwd, package/build scope, loading,
analysis/output bounds and caller uncertainty. Full JSON availability is not
exposure. State uninspected boundaries, elapsed/budget/reporting reserve and
what would decide open questions. No category/function/family/finding quota or
detection-effectiveness claim; a valid outcome is no demonstrated violation.
