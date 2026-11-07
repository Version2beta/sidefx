# Maintainer validation

The skill is standalone. From a copy of its directory, run the four stdlib helper
checks (the real-CLI check skips unless explicitly enabled):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -v
```

For a real-CLI procedural dry run, build the current sidefx checkout offline,
then run the tests from this directory with an absolute binary path:

```sh
# Run the build from the sidefx checkout, not the skill directory.
evidence=$(mktemp -d /tmp/sidefx-skill-validation.XXXXXX)
env GOPROXY=off GONOPROXY=none GOSUMDB=off GOTOOLCHAIN=local GOFLAGS=-mod=readonly \
  go build -o "$evidence/sidefx" ./cmd/sidefx
# Then, from this skill directory:
PYTHONDONTWRITEBYTECODE=1 SIDEFX_BIN="$evidence/sidefx" \
  SIDEFX_EVIDENCE="$evidence" python3 -m unittest discover -s tests -v
```

The optional check creates and **retains** a fresh stdlib-only temporary Go
module. It runs no arbitrary target application. Its authored contract says a
replacement must leave destination contents unchanged on error; the injected
implementation writes first and then rejects a denied replacement. This is
**author-known test data, not blind defect discovery**. The prospective record
precedes the harness's implementation inspection, not authorship or content
hashing. Location is explicitly unresolved at selection, then verified from the
source. No production impact or novelty claim follows.

The harness retains actual function-view stdout, stats stderr, exit statuses,
compact JSONL, binary hash/build info, Go version, explicit isolated environment,
command/read chronology, selection, source, probe and conclusions. It verifies:

- direct-marked versus inherited-marked established categories remain distinct;
- real CLI candidate-only `possible_effects` survive reduction without promotion,
  and actual unresolved uncertainty stays separate;
- missing locations/reasons/provenance remain unknown;
- `-all` exposes an omitted empty summary without a purity label;
- exact local function matching and substring package matching work;
- the named probe observes denial **and changed bytes**, demonstrating the
  injected contract violation. A separate lower-level-write-error concern is
  explicitly untested. Original fixture source hashes remain unchanged.

The fixture has no app-specific storage or initializer. HOME, XDG config/data/
cache/state, TMPDIR, GOCACHE and GOMODCACHE are isolated, with offline readonly Go
settings and workspace disabled for this standalone fixture. The only runtime
file writes tested are under `t.TempDir()`. No listener or external service is
used. These isolation assumptions do not transfer automatically to target apps.

## Checked environment and results

### Historical validation before the sidefx rename

Implementation validation used Reffector working tree HEAD
`95542cdc22d708a7c87dc604d7323ec364fdd837` (pre-existing dirty documentation and
untracked artifacts), Go `go1.27.2 darwin/arm64`, and Python 3.15. No analyzer/CLI
changes were required. The real output emitted six of seven functions: four
established-effect functions, one inherited-effect function, three direct-effect
functions, one uncertain function, and one omitted function. The possible-only
function retained `fs.read` / `fs.remove` with no established effects and false
uncertainty; the open callback retained true uncertainty. Direct/inherited counts
are not a partition. Five unittest checks passed, including the real-CLI dry run
and its isolated Go probe. Malformed input failed nonzero without partial rows.

Pi 1.4.0's installed discovery/frontmatter APIs were checked without starting a
model/server, installing a package, or changing settings. Project discovery
requires project trust; user overrides/collisions can still affect availability.
Use `/reload` and `/skill:sidefx-function-first` in the actual host session;
interactive invocation itself is not exercised by these checks.

Local raw evidence for this implementation is retained under
`/tmp/reffector-skill-validation.4E5VMh/`; it is disposable and not bundled. This
small fixture checks compatibility and evidence discipline, not discovery quality
on large real projects, closure location mapping, all build variants, or runtime
safety of arbitrary target tests. The helper loads its array in memory and is
not a streaming analyzer or scoring system.

## Complete operation orientation (subsequent repository workflow)

`compact.py` remains a function-array reducer; it is not a category/site index.
The installed native `sidefx index` and `sidefx acquire` need no Python or source
checkout; saved projection also needs no Go. This separately copied skill does
not supply the executable. `sidefx skill install` installs the canonical guide plus
full59/recognition/acquisition references under `~/.agents/skills/sidefx/references/`.
The old Python operation index is development/reference-only, not required for
normal installed use. See the installed `references/operation-index.md` there for fresh acquisition provenance/no-loss checks,
legacy-input limits, distinct counts/overlap and exact-category/ID pagination.

Repository checks (no arbitrary target execution):

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tools -p test_operation_index.py -v
go test ./cmd/sidefx -run 'Test(OperationIndexCLI|InstalledIndex)' -count=1 -timeout=180s
```

These exercise the built CLI to saved-JSON index/retrieval with >10 categories,
a singleton and >24-site category, established/possible-only labels, selected and
dependency origins, unresolved separation, stable bytes, filter/scope honesty,
capped `-all`, full pagination, caller candidates and contracts. Unknown origin,
malformed input and hash/cursor guards have native tests. Completeness of a declared
fresh query additionally requires the acquisition record and no site-output loss;
complete emitted legacy records alone do not establish selected-scope orientation.
This is procedural/serialization evidence, not blind defect-discovery validation.
