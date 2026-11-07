# Complete operation index and bounded investigation

## Mechanism and decision

`sidefx index` is a native Go stdlib projection of **saved effect-view schema 1
JSON**. It never loads Go packages, runs SSA, recognizes APIs, ranks leads,
consults issue locations or samples sites. The standalone installed binary works
without skills, Python, Go or a source checkout for saved-report processing.
Fresh `sidefx acquire` analysis requires Go 1.26+ on PATH and authorized cached
dependencies. Function-view arrays and audit JSONL are rejected by the index.

Existing normal CLI output remains unchanged. The additive reserved first command
names `index` and `acquire` select these workflows (use `./index` or `./acquire`
for same-named package directories). No recognition, propagation, site IDs,
analyzer schema, caller queries or normal analysis defaults changed. A focused
native saved projection avoids repeating package loading and preserves unknown
record fields; it does not duplicate production recognition. The prior
`tools/operation_index.py` remains **development/reference-only**, with optional
semantic parity checks, not a supported installed runtime dependency.

`sidefx skill install` installs the canonical guide and **all three local
references**: this contract, the full [59-family matrix](operation-coverage-audit.md)
and [recognition decisions/limits](operation-recognition-decisions.md). Their
installed bytes come directly from the authoritative docs embedded in the binary;
no generated or second catalog is maintained. Skills are optional. Neither
installation nor saved indexing requires Go, Python or a checkout.

The traced pipeline is exact API/SSA recognition (`effects.go`, `operations.go`,
`analyzer.go:build/ssaOperation`) → source-position merging and operation evidence
(`details.go:addSiteEvidence`) → established/candidate propagation → root relevance
and site/filter/graph budgets (`effectview.go:effectReport/reverseGraph`) → groups
and JSON → native saved index → small exact-category/ID pages → source/caller investigation.
Function direct/inherited marks are summaries, not origin counts. A Site's established
label can be reached from a particular caller only via a candidate edge; the projection
does **not** upgrade that path. Reverse source resolution, candidate flags, generic
variants, contracts and uncertainty are copied, not inferred.

## Phase A: acquire and read the entire category index

First establish target/source/producer identities, dirty source manifest, allowed
reads/writes/runtime actions, explicit package/build scope and budgets. Prefer
qualified cached JSON: projection and pagination require **no additional SSA**.
Legacy JSON without recorded command/build provenance has a complete emitted-record
index, **not an established complete selected-scope orientation**. Full JSON on disk
is not model exposure. Record every category ID and coverage/bounds as presented;
do not put all operation records in model context.

### Fresh acquisition with the installed binary

First locate the executable: `command -v sidefx`. If absent, stop and obtain an
authorized installed executable; skill installation does not install the binary.
Either add its verified directory to PATH or use an explicit absolute path for
**every** command (for example `/opt/sidefx/bin/sidefx index REPORT.json`).
Do not build from an arbitrary target checkout or install/download as a fallback.
Use `sidefx acquire -help` and `sidefx index -help` to validate the binary surface.

From the authorized target cwd, choose a **new nonexistent run-directory** beneath
an existing authorized external parent. Do not pass an existing directory, even
if empty. This example chooses **direct selected roots**, excluding dependency
origins and candidate-only helper reachability. Omit `-direct` only when the
explicit scope is reachable origins, including candidate paths. Neither scope
covers every loaded dependency body. Preserve workspace, tags/platform and tests
settings rather than silently changing them. Source revision and relevant dirty
identities remain caller-supplied evidence and must be retained separately.

```sh
export GOPROXY=off GONOPROXY=none GOSUMDB=off GOTOOLCHAIN=local GOTELEMETRY=off
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly"
# Set EVIDENCE to an authorized NEW nonexistent path before invoking.
# Preserve GOPRIVATE, tags and the approved workspace; GOWORK=off only for approved standalone modules.
# GONOPROXY=none disables private download bypass without changing GOPRIVATE.
sidefx acquire -out "$EVIDENCE" -direct -max-depth 0 -max-sites 100000 -max-edges 1 ./...
sidefx index -provenance "$EVIDENCE/acquisition.json" "$EVIDENCE/effects.json"
```

`acquire` refuses run-directory collisions **before analysis or writes**, and uses
exclusive creation for acquisition-start.json, effects.json, effects.stderr,
acquisition.json and index.json. It invokes the existing analysis pipeline once
with explicit unfiltered `-view effects -json -stats`, records actual argv/cwd,
executable SHA-256/build info, allowlisted build environment and effective Go
settings, exit and report SHA-256, then creates the native index without SSA.
The bounded environment record includes GONOPROXY so the enforced private-bypass
policy is visible. GOPROXY URL userinfo/query parameters (or malformed proxy URLs)
redact the whole recorded GOPROXY field explicitly; ordinary credential-free
settings, off and direct remain unchanged. No secret-derived hash is substituted.
This is narrow protection for known proxy credentials, not universal secret
sanitization: environment/provenance remains bounded caller-authorized evidence.
Saved raw records and caller-supplied provenance are preserved, not scrubbed.
GONOPROXY=none prevents inherited private patterns from bypassing GOPROXY=off;
these Go settings are not proof of a network sandbox.
It accepts direct/tests/exact/all, positive site/edge caps, nonnegative depth and
paired positive dependency-VTA budgets; no filter/contracts/unresolved options.
Normal CLI still supports those options separately. Numeric defaults are 100000
sites, 1 edge and depth 0; they are recorded, not assumed unlimited. Nonzero exit
on analysis failure or incomplete site output is a STOP, not zero inventory;
incomplete successful reports retain an honest index. Read diagnostics and bounds.
Permission/I/O failure may leave partial evidence; use a separately authorized new
run rather than overwriting or retrying in place. A successful acquisition indexes
recognized operations, **not the omitted unresolved inventory**.

Stop on loading/provider/permission blockers; do not repair dependencies or retry
through another route. If the **ordinary output cap** was inadequate, record the
incomplete result; only separately permitted re-acquisition with an adequate cap
and unchanged scope can close that gap. It reloads/rebuilds analysis and has cost.
`-all` merely adds unfiltered **function inventory**; it does **not** bypass MaxSites
or MaxEdges and cannot repair a capped effects report. There is no unlimited-origin
flag. Depth zero deliberately leaves callers incomplete without losing origin
category visibility. Later site queries reload/rebuild Go analysis; the native projection
instead reparses the saved JSON on each invocation (no persistent cache).

A hash-bound `-provenance` record requires successful exit, nonempty producer,
build and cwd fields, and the actual argv array with explicit `-view effects -json`,
positive `-max-sites`, explicit package patterns and **no** effect/API/site/operation/
reason/package/function/unresolved/contract filters. Supported analysis settings are
`-direct`, `-tests`, `-exact`, `-all`, `-stats`, depth/edge limits and paired dependency
VTA budgets. Unsupported/duplicate/filter flags prevent a complete-orientation
claim (the emitted index is still usable). The validated, hash-bound record is
retained once at `bounds.acquisition` independently of that eligibility, including
filtered/unsupported argv, cwd, producer/build and recorded caps. `bounds.query`
is only the eligible unfiltered interpretation; rejection reasons remain explicit
in `orientation_limits`, without calling supplied command settings unavailable.
Malformed provenance or hash mismatch
fails without partial stdout. Provenance is caller-supplied evidence, **not
cryptographic proof of execution**. A forged or lossily rewritten producer output
cannot be recovered or authenticated by this projection.

### Index/count/completeness contract

- `categories` contains **every exact identifier present in emitted Site.effects or
  Site.possible_effects once**, lexicographically sorted. Site labels are authoritative:
  a CLI `-effect` can suppress co-labels in `groups`, but not in the retained Site.
  There is no category shortlist, frequency threshold, first-24 rule or sample.
- Each row's `sites` counts distinct canonical IDs. `established` and `possible_only`
  partition membership **for that category**, not execution/reachability from every
  caller. A site possible-only for one category can have another established label.
  Runtime conditions use established labels plus uncertainty, not candidate labels.
- `selected_origin`, `dependency_origin`, `unknown_origin` partition those sites using
  matching reverse **origin** nodes' `selected` flags. Any matching selected origin
  wins when variants merge. Missing matching origin evidence is unknown, not zero
  dependencies. This is not the function direct/inherited distinction, nor a
  candidate-caller classification. `uncertain` overlaps established/possible counts;
  it is a site flag, not the number of unresolved behaviors or a defect score.
  Per-category `unresolved` counts schema-unresolved membership (no established
  site labels and uncertain), overlapping that category's possible-only count.
  `uncategorized_other_records` separately retains the count of any input records
  with neither category nor uncertainty (for example auxiliary contract records).
- Categories overlap: `category_site_memberships` is not the distinct-site total.
  `categorized_distinct_sites` and `multi_category_sites` make this explicit.
- `bounds.emitted_records=complete` guarantees no projection sampling/loss **within valid
  input**, even if producer output was capped. `declared_query_categories=complete`
  additionally requires qualified unfiltered acquisition provenance, equal matching/
  emitted counts and no site truncation. Without these it is `not_established`.
  This means all present labels in that declared, build-qualified analyzer scope,
  not all Go behavior, all builds, all dependencies or absence of defects.
- Original coverage and site/edge/depth truncation are retained independently.
  Dependency-VTA/context budgets and truncation, when exposed in coverage, remain
  visible under `dependency_analysis_bounds`; default discovery remains bounded.
  Omitted optional analysis counts are unavailable, not inferred zero.
- `unresolved_view_records` preserves the schema's effect-free uncertain-site set,
  which **can overlap possible-only categories**. `uncategorized_unresolved_records`
  counts only uncertain sites with neither established nor possible labels;
  `possible_only_unresolved_overlap` states the difference. Identified unresolved
  count is not a retrievable unresolved inventory: normal operation JSON usually
  omits those records. Obtain a separate authorized `-unresolved` acquisition (or
  audit export) if needed; don't merge incompatible scopes or silently add its
  records to the original index. The native index supports effect-view unresolved JSON,
  not audit JSONL. `-uncategorized-unresolved` pages the truly uncategorized subset.
- Schema 1 does not expose the complete registered vocabulary or all filters/build
  settings. `registry` is unavailable: an absent category is **not** recognized-zero
  and is not proof the analyzer does not recognize it. No registry is hardcoded.

The native projection validates schema, unique IDs/labels, coverage counts, group membership,
reverse references (including each referenced edge's endpoints inside that reverse
graph's own node refs) and truncation before emitting output. It cannot detect losses
that a rewritten input hides by falsifying all corresponding metadata. Unknown
fields in retrieved records remain intact; no API/evidence/classification rewriting.

The minimum graph schema requires node `package`/`function` strings (empty strings
are valid for synthetic symbols), boolean `selected`, and a nonnegative integer
`depth` on every reverse-node reference. Missing/mismatching origin evidence can
honestly yield unknown origins; malformed identity types cannot. For legacy saved
records, node variant/uncertainty/synthetic flags, edge candidate/synthetic/source,
and reverse revisit/cycle fields may be absent. When present their known types are
validated (including source symbols/location); source may also be null. These
optional legacy fields are not filled in or inferred. Unknown fields and numeric
precision remain preserved. The Python development reference is a valid-input
parity oracle, not the installed validator's malformed-input authority.

## Phase B: small prospective selection and focused retrieval

Read **every index category ID and its bounds**, not every site. Choose a small
prospective portfolio across the index by consequential local obligations and
tractable evidence, not frequency/rank/importance by count. Record one alternative
lead. Consult only relevant parts of the supplied identity-qualified **full 59-family
reference**, not a regenerated catalog or exhaustive checklist. If the required
reference/authority is unavailable, stop for it. An index is an orientation aid,
not a category investigation quota.

Retrieve exact categories in small pages; use the emitted `next_offset` until null
when the full corresponding set is needed. `total` is the complete corresponding
set **in that saved report**, not all source operations. For example:

```sh
sidefx index -category fs.write -limit 8 "$EVIDENCE/effects.json"
# Copy the returned next_cursor; it binds input bytes, selector and next offset.
sidefx index -category fs.write -limit 8 -cursor CURSOR "$EVIDENCE/effects.json"
# Alternatively use next_offset with input_sha256; both must come from that set.
sidefx index -category fs.write -limit 8 -offset 8 -expect-sha256 SHA256 "$EVIDENCE/effects.json"
sidefx index -site SITE_ID "$EVIDENCE/effects.json"
```

All flags precede REPORT.json; one exact, nonempty selector is required for pages.
IDs are sorted lexicographically; offsets are stable for the same input hash.
`next_cursor` is null at completion; otherwise pass it unchanged with the **same
selector**. Changed input, wrong selector, malformed cursor or cursor+offset fails
without partial stdout. It is a consistency token, not authentication. Explicit
offsets alone are not bound: use `-expect-sha256` on subsequent offset pages.
Use `-uncategorized-unresolved` for the truly uncategorized subset.
Pages have 1–100 sites (default 10), no silent drop of later records. Absence or
invalid offsets fail explicitly. Optional provenance applies to pages too. A page
copies full sites, corresponding reverse entries, referenced nodes/edges and
interface contracts, including all uncertainty/evidence labels. Graphs retain
producer bounds; site limits do **not** bound each source graph's bytes. Use a
small site count and bounded source graph to fit context; raw JSON memory/output
is not streamed or byte-capped. For additional reverse callers not saved, use
`sidefx -site ID -max-depth 2 -max-sites 1 -max-edges 50 -json SCOPE`, retaining
`-unresolved` for unresolved-only origins and original build/package settings.
That query repeats analysis and may expose candidate callers; it does not upgrade
them. Package/function filters can hide callers. Source locations/IDs can change
with source; don't mix revisions or substitute new IDs into historical reports.

Before implementation reading, record site/API/category/location, origin and caller
uncertainty; why it matters; provisional obligation; falsifiable hypothesis plus
strongest protective explanation; minimum next source/contract evidence and budget.
Do not let output order or known issue locations choose index/samples.

## Phases C–E: bounded follow-through, contract and chronology

**C.** Follow the selected operation's controlling decisions/mode and nearest relevant
callers/callees. Choose the most consequential unanswered connection, not all of them:
persistence **and recovery-source retirement order**; accesses to the **same shared
object** and caller-lock assumptions; borrowed buffer retention/copy versus producer
reuse (first check retention is enabled); caller error/skip/retry/fallback disposition;
queued remainder and consumed/committed/restart positions **after rejection**, including
one next iteration; cleanup/deletion representations and what survives reconstruction.
No architecture sweep, universal error/lock analysis or separate vulnerability brainstorm.
Stop/park when the next boundary exceeds the local budget.

**D.** Read the nearest contract/interface/comments, relevant tests and protective
branches. Separate authority from inferred expectations and observed behavior. Seek
counterevidence and mode gates. Only permitted focused verification may execute;
source of a test is not a run. Distinguish **confirmed contract violation**,
**reproduced behavior, contract uncertain**, **source-supported hypothesis**,
**rejected hypothesis** and intentional tradeoff. Missing contract or safe verification
is a productive stop, not permission to strengthen claims or retry a provider block.

**E.** Log identified versus index-presented categories versus sites actually retrieved,
leads selected before reading, source ranges, actual subquestions, contract reads,
commands/statuses/tests, evidence for/against, productive stops and time/reporting
budget. Full inventory availability is not exposure; counts are not lead quality.
Reserve reporting time and an alternative lead; no finding/function/family quota.

## Validation and limits

Native unit and CLI tests cover full category/index semantics, malformed inputs,
unknown-field fidelity, origin/possible/unresolved distinctions, closed graphs,
provenance/hash/cursor guards, complete pagination, filters and capped `-all`.
Installed integration tests copy the binary to an unrelated directory, install
with temporary HOME, verify every embedded reference byte/link and execute the
documented acquire/index commands with Go but **no Python** on PATH. Saved
retrieval and installation also run with neither Python nor Go available.

```sh
# Authorized offline repository validation; not target runtime authority:
go test ./internal/operationindex ./internal/skillinstall
go test ./cmd/sidefx -run 'Test(OperationIndexCLI|InstalledIndex)' -count=1 -timeout=180s
# Optional development reference checks only (not an installed requirement):
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools -p test_operation_index.py -v
SIDEFX_INDEX_PYTHON_PARITY=1 go test ./cmd/sidefx -run '^TestOperationIndexCLI' -count=1 -timeout=180s
```

The historical NATS regression is **saved JSON processing only**, not a new analysis,
test, issue search, reproduction or detection-effectiveness evaluation. Original
producer identity is separate from projection implementation identity. Costs must report actual serialized
bytes/categories/wall time and process RSS (not incremental memory unless actually
measured). Deterministic execution is not model-token cost; no token estimate is a
measurement. Historical regression and run-specific preservation/cost evidence live
in the authorized external evidence directory, not in target code or bundled snapshots.
