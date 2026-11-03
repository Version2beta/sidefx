---
name: sidefx-function-first
description: Investigate observable Go correctness, reliability, resource-lifecycle, or security behavior bugs by selecting functions from sidefx summaries before reading implementations. Use for function-first defect discovery, not effect inventories; verify leads against contracts, source, and authorized isolated tests.
compatibility: Go 1.26+, a verified sidefx binary supporting -view functions, and Python 3 for the optional bundled summary helper.
---

# Function-first investigation

Find real behavior defects, not lists of effects. A summary is a selection aid,
not a defect, execution trace, risk score, or proof of safety. Work on **one
function at a time**. Do not impose a finding count, exhaustive graph, fix, or
refactoring. Investigation alone does not authorize edits or test execution.

## 1. Establish scope and retain evidence

Default to read-only target inspection and offline loading. Preserve dirty and
untracked work. Do not repair dependencies, download, switch revisions, or
change build/workspace settings without authority. Use a fresh temporary or
user-approved evidence directory outside target source; report its actual path.
Record UTC command/read chronology, cwd, scope, source revision and dirty state
(including relevant untracked content identity), binary path/hash/build revision,
Go version, build tags/flags, workspace/replacements, OS/architecture and whether
tests are included. Hash or retain relevant dirty source, not just HEAD; record
unknown provenance explicitly. Never read sensitive material just to inventory it.

Verify `sidefx -help` against the actual binary. If building from a sidefx
checkout is authorized, use `go build -o <evidence>/sidefx ./cmd/sidefx`
there, recording environment, stdout/stderr and exit status. Do not build from
an arbitrary target module or assume a preinstalled binary matches source.

From the target module/workspace, start with the actual function view:

```sh
# Keep existing build settings; do not silently drop tags or workspace context.
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-mod=readonly"
evidence=$(mktemp -d "${TMPDIR:-/tmp}/sidefx-functions.XXXXXX")
# Record the resolved binary, provenance, cwd and command before running it.
printf '%s\n' 'sidefx -view functions -json -stats ./...' > "$evidence/command.txt"
date -u '+%Y-%m-%dT%H:%M:%SZ' > "$evidence/started.utc"
if sidefx -view functions -json -stats ./... \
    > "$evidence/functions.json" 2> "$evidence/functions.stderr"; then
  status=0
else
  status=$?
fi
printf '%s\n' "$status" > "$evidence/functions.exit"
test "$status" -eq 0  # Stop on failure; preserve diagnostics, not a zero-effect claim.
```

`./...` is recursive; use a different explicitly recorded scope only when the
user requested it or the scope needs approval. Function summaries cover selected
source functions, not all loaded dependency functions. `-tests` includes test
functions without executing them. Missing cached modules/toolchains or failed
loading are blockers, not permission to download or silently narrow scope.
sidefx loading is preferable to raw `go list` preflight in workspaces: raw Go
tooling can update workspace sums. Retain raw JSON, stats **stderr**, errors and
status even when analysis fails. Check successful loading and selected package,
module and build coverage in stats; independently record provenance that stats
cannot provide. Static loading does not execute the target's initializers.

## 2. Reduce without losing evidence distinctions

Use programmatic reduction, not a giant model dump. Optional stdlib helper
(relative to this skill directory):

```sh
python3 scripts/compact.py "$evidence/functions.json" > "$evidence/compact.jsonl"
```

Inspect small explicitly filtered slices and coverage totals, retaining the raw
records for each choice. Standard `-view functions -json` records contain
`package`, `function`, `effects` (each `identifier` and `direct`), optional
`possible_effects`, and `uncertain`. The helper partitions established effect
identifiers by their `direct` mark into `direct_marked` and `inherited_marked`,
copies `possible_effects` (defaulting an omitted field to `[]`), and copies
`uncertain`. It emits `null` for `location`, uncertainty reasons, and uncertainty
provenance because they are absent from this input; it does not drop those fields
from standard function records. The distinct `-details` projection has
per-function direct/inherited uncertainty flags and source-located uncertainty
sites/reasons. `-audit-unresolved -all` function records have owner `locations`.
These are separate schemas: use them or a location-only AST lookup when
relevant; do not guess or merge their fields into the compact summary. The
ordinary function-view schema has no source location; a function identity may
map to multiple declarations or variants. The current gap is that locations and
uncertainty details do not share one summary row, not that the CLI lacks them;
no CLI change is needed for this workflow. If a future consumer needs a
single-row projection, a minimal additive output change would be an optional
`locations` array and direct/inherited uncertainty flags, with reasons
referencing the existing source-located sites. Malformed input is an error, not
an empty summary. The helper reads the JSON array in memory; it does not rank
functions or establish locations.

Interpret the current schema carefully:

- `effects[].direct:true` marks a recognized category present directly in the
  function. `direct:false` means it is reached through established calls with
  no direct occurrence. If a category is both direct and inherited, the record
  marks it once as direct; `inherited_marked` is not a complete list of inherited
  paths. Do not treat these marks as disjoint path counts or infer exclusive
  ownership.
- `possible_effects` depends on candidate relationships, not established calls.
  `uncertain` is a separate boolean; neither its reasons nor local versus
  inherited origin is available here. `false` does not establish complete knowledge.
- Stats' direct/inherited/uncertain function counts overlap. Emitted/omitted
  counts describe this projection, not source coverage or runtime execution.
  Optional supplementary `-view functions -all -json -stats ./...` exposes omitted
  summaries for controls; retain its outputs separately. Empty or omitted
  summaries are **not pure/safe**, even with no recorded uncertainty.

Meaningfully consider functions with few/no direct marks but consequential
inherited or possible behavior: a caller's ownership, policy or lifecycle
obligation can matter more than the operation-owning helper. Do not select only
large summaries, many effects, output order, familiar symptoms, or direct sites.
There is no effect-count ranking or risk score.

## 3. Select prospectively, then inspect

Before opening a candidate's implementation, locate its declaration using a
focused site query or a location-only Go AST lookup. Do not review bodies during
location lookup. Function JSON has **no file:line**; methods, closures, generic
and test variants require verified mapping. An inherited site's location belongs
to its operation owner, not necessarily the selected caller. If the caller's
location is unavailable, record it as unresolved rather than guessing. An
effect-free owner may have no site.

Save a brief, timestamped selection record **before the first implementation
read**. Keep it concise and evolving: only the pre-inspection fields below are
required before opening the body. Fill later fields as evidence arrives; leave
unavailable details explicitly unknown rather than guessing.

```markdown
### Candidate: <package.function>

**Before implementation read**
- Selected: <UTC timestamp or chronological position>
- Location: <verified file:line, or unresolved and how it was sought>
- sidefx evidence: <direct-marked; inherited-marked; possible_effects;
  uncertain; query/record reference, or not used/unavailable>
- Prior source, symptom, or contract knowledge: <state it; do not backfill>
- Prospective obligation/question and rationale: <contract, ownership, or
  security concern; what behavior could violate it; why this function; what
  evidence could confirm or reject it? Mark authority provisional if unknown.>

**During investigation — add only what is relevant**
- Trigger and state: <input, configuration, prior state>
- Expected behavior: <authoritative contract and citation, or inferred
  expectation and rationale; label which>
- Source path: <caller/callee, validation/state changes, side-effect site;
  distinguish source inference from runtime observation>
- Observed behavior: <test/observation and actual result, or not reproduced>
- Unverified hypothesis: <behavior not yet observed>
- Side-effect target and relevant constraints: <operation/resource and bounds>
- Ownership, policy, authorization, cancellation, retry, cleanup, and errors:
  <relevant subset; security boundaries only when evidence supports them>
- Evidence for / against: <source, test, output, or none yet>
- Remaining unknowns: <conditions not established>

**At resolution**
- Outcome: <confirmed defect | reproduced behavior, contract uncertain |
  unverified hypothesis | rejected hypothesis | intentional behavior>
- Supporting observations/tests: <what was exercised and result>
- Assumptions and untested production conditions: <limits>
```

A prospective question is a hypothesis, not a finding. Keep authoritative
contracts separate from expectations inferred from source/design and behavior
observed in tests. A mismatch with an inferred expectation alone is not a
confirmed contract violation. Disclose prior knowledge; do not backfill a
prospective story. Names/effects suggest questions, not behavior.

Use bounded location/evidence queries only as needed, copying identities from
output. For example, with `pkg` and `fn` taken from the selected record:

```sh
sidefx -json -package "$pkg" -function "=$fn" \
  -max-depth 0 -max-sites 20 -max-edges 40 ./...
# For effect-free unresolved origins, repeat with -unresolved.
```

`-function '=NAME'` matches the exact local function identity, **not** its
package-qualified concatenation. `-package` is substring matching, not exact;
verify the returned package and variants. Function filters without `=` and
`-effect` are substrings. `-api` is an exact API identity; `-operation` is an API
substring (effect view or compatibility details only). Summary-view graph limits
are invalid; switch to the effect view for sites/reverse callers. Filters can
hide callers. Depth 0 shows origins; raise to 1 or more only to answer a question.
Check report and per-site truncation and site/edge/depth budgets. Budgets bound
output, not loading cost. Avoid `-details` over an entire large graph and do not
build an exhaustive graph by default. Candidate edges and `-contracts`
method-set-compatible implementations are leads, never established execution.

### When a call target is unresolved

An empty effect- or API-filtered result means only that no matching item was
reported within that query's selected packages, filters, and bounds. It does not
prove the operation or behavior is absent. If an unknown target matters to the
hypothesis, use the existing unresolved view with the relevant reason:

```sh
# Open function-value / callback target
sidefx -unresolved -reason function_value -package "$pkg" \
  -function "=$fn" -max-depth 0 -max-sites 100 -max-edges 100 ./...

# Open interface target; -contracts adds compatible implementation leads
sidefx -unresolved -reason interface_dispatch -contracts -package "$pkg" \
  -function "=$fn" -max-depth 1 -max-sites 100 -max-edges 100 ./...
```

These use the default effect view. `-unresolved` selects uncertain sites with no
established site effects; possible effects may still be present. For a
recognized-effect site that also has uncertainty, omit `-unresolved` and use
the default view with `-reason`. `-contracts` returns leads, not established
calls; it does not apply to function-value dispatch. There is no separate
`-function-value` flag; use the `function_value` reason. Recheck package coverage
and truncation, then verify relevant calls in source. An empty unresolved result
is not proof that no unknown behavior exists. Use these queries only when the
hypothesis depends on the boundary, not for exhaustive call-graph exploration.

Read the selected implementation and only the callers, callees, contracts,
interfaces, and tests needed to distinguish the question. Trace behavior around
the selected effect, not just the reported operation. When relevant, follow
validation and policy decisions; state changes and invariants; resource ownership
and transfer; success, failure, and partial-completion cleanup; retries;
cancellation, deadlines, and concurrency; and error propagation or recovery.
These are investigation prompts, not mandatory checks for every function.
Record read ranges and commands in chronology. Establish intended behavior from
authoritative contracts, not current behavior alone; label expectations inferred
from source/design as inferred. Form falsifiable hypotheses and record evidence
that supports or contradicts them. For security-relevant leads, document
attacker control, trust boundaries, authorization, and validation only when
supported by source or contract evidence; a missing check in one function alone
is not a vulnerability. Separate analyzer evidence, source observations, runtime
observations, and inference.

If the original question is rejected and source suggests a different concern,
record a **source-directed transition** and new question before following it.
Do not present that follow-up as an original summary-only discovery. Finish or
park this function honestly before selecting another prospectively.

## 4. Check safely and report plainly

Tests/probes require authorization beyond read-only analysis. Before executing
anything, inspect package initializers, test setup and invoked paths for side
effects. Prefer a disposable source copy and the smallest named test; do not run
an arbitrary target app or broad suite. Isolate HOME, temp, config, data, cache,
state, autosave and storage paths under a fresh root, including app-specific
paths/default globals. Preserve needed offline dependency-cache access explicitly.
Use loopback only if networking is authorized, never external/production services
or real credentials/data. Bound time, resources and concurrency. If isolation
cannot be established, stop and report the untested question. Do not fix source
automatically; any edits need separate authority.

Produce a plain report: provenance and exact commands/status; coverage/omissions
and truncation; candidate records and source-directed transitions; contract,
source/call and test evidence; observable impact and limits; and the stopping
record below. Call a behavior a **confirmed defect** only when an authoritative
obligation and violating observation are both established. If behavior is
reproduced but the contract is uncertain, label it **reproduced behavior,
contract uncertain**. Otherwise distinguish **unverified hypothesis**,
**rejected hypothesis**, and **intentional behavior**. A mismatch with an
inferred expectation alone is not a confirmed contract violation. A passing
probe supports only the behavior its assertions exercise; say what it proves.
A valid outcome is no demonstrated defect. Novelty is **unknown** unless
separately checked after discovery with authorization; do not import issue
history or previous findings as discovery hints.

## 5. Stop with an evidence-based decision

Before completing the report, add a concise stopping record:

```markdown
- Elapsed / time or budget remaining: <actual values, or not set>
- Functions and behavioral hypotheses investigated: <brief list>
- Credible leads still open or parked: <list with reason, or none identified>
- Stop reason: <budget limit | no credible leads remain in reviewed scope |
  validation blocked | diminishing expected value | other concrete constraint>
- Basis: <evidence/review supporting this decision>
```

Finding a defect or finishing the report is not, by itself, a stop reason. Do not
claim that no leads remain merely because no more were selected. The record makes
the decision auditable; it does not require exhaustive inventory review, a
minimum number of functions/findings, or consuming the full budget. Stop early
when remaining credible leads are unlikely to justify further work, and say why.

## Invocation and reuse

Pi discovers this project `.agents/skills/sidefx-function-first/SKILL.md`
when the project is trusted (subject to host overrides/name collisions).
Invoke `/skill:sidefx-function-first <scope and authority>`; use `/reload`
after adding/editing it in an active session. For reuse across projects, review
and copy (or symlink) the **whole standalone directory** to
`~/.agents/skills/sidefx-function-first`, without overwriting another skill.
Do not install globally without permission. Other Agent Skills-compatible hosts
can load the same directory; no SDK, settings change or new CLI is required.

Maintainer checks and limitations: [validation.md](references/validation.md).
