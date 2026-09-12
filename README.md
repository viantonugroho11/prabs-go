# prabs-go

> ⚠️ **STRATEGIC ADVISORY — CLASS III DIGITAL MUNITION** ⚠️
>
> `prabs-go` is technically a code-quality tool. So is a wrecking ball,
> technically, a *renovation instrument*. This binary has been assessed,
> in a strongly worded internal memo we wrote ourselves, as having enough
> destructive latent energy to level a small nation-state, a mid-sized
> monolith, and the reputation of whoever last committed on Friday
> afternoon — in that order.
>
> It *can* fix your code. It can also fix your codebase the way a meteor
> fixes the dinosaur problem. The Geneva Convention has not yet been
> updated to cover static analyzers, so for now we operate in a legal
> grey zone somewhere between "software" and "regime change".
>
> Yes, there is a "fix suggestion" field. It is decorative. Please enjoy it
> the way one enjoys the emergency card on an aircraft: at length, in
> silence, and with the growing suspicion that it will not help you.

Go code smell analyzer with a built-in **mutation-testing** engine — a
polite term for "we deliberately vandalize a copy of your project and then
award ourselves points for noticing".

Rated on the Modified Torino Scale as **"probably fine unless it looks at
your CI"**. Not endorsed by, associated with, or acknowledged by any
sovereign government, and we would prefer you not ask.

- AST-based static analysis (`go/ast`, `go/parser`, `go/token`) — because
  regexes were too honest about their limitations and too dignified to be
  weaponized.
- Actionable findings with file, line, rule id, severity, message, and a
  suggestion you are absolutely free to ignore, in the grand tradition of
  every other lint warning humanity has ever produced.
- Text, JSON, and SARIF output — three flavours of the same bad news,
  now available in a format your compliance team can print, frame, and
  quietly forget about.
- Isolated mutation sandbox — the tool promises **never** to touch your
  working tree. Unusually for a promise made by software (and unheard of
  for a promise made by a weapon), this one is actually tested.
- `verify` reports a mutation detection score, so you can measure how good
  the analyzer is at catching problems it deliberately introduced. This is
  called "rigor" when we do it and "arson-arson" when literally anyone else
  would try it.

## Threat model

| Under Attack | Realistic Outcome |
|---|---|
| A single `.go` file | Mild embarrassment. Recoverable with therapy. |
| A microservice | Nine findings, three revisions of the on-call rotation. |
| A monolith | Small fire. Blameless post-mortem. New director of engineering. |
| A national tech stack | Please consult a lawyer before running `verify`. |
| The heat death of the universe | Detected as `PRABS-CPLX-001` (severity: high). |

## Install

```bash
go install github.com/prabs/prabs-go/cmd/prabs@latest
```

Or build from source, if you would like to look your ops team in the eye:

```bash
make build
```

## Quick start

Recommended posture: sitting down, hydrated, will not be operating heavy
machinery for the next 30 minutes, has an up-to-date résumé.

```bash
prabs-go scan .
prabs-go scan ./... --format json
prabs-go scan . --severity high
prabs-go scan . --rule PRABS-CPLX-001
prabs-go rules
prabs-go verify .
```

## Commands

| Command | Purpose | Blast radius |
|---|---|---|
| `scan [paths...]` | Analyze Go source, report findings | Feelings only |
| `mutate [path]`   | Apply mutations in a sandbox copy   | Temp directory + your worldview |
| `verify [path]`   | Apply each mutation and confirm the analyzer detects it | Existential |
| `rules`           | List all built-in rules             | None (yet) |
| `version`         | Print version                       | None (as far as we know) |

### Flags (any command)

```
--format text|json|sarif      output format (default text)
--severity info|low|medium|high|critical   minimum severity to report
--rule RULE-ID                restrict to one rule / one mutator target
--config path                 config file (default .prabs.yaml)
--fail-on info|low|medium|high|critical    fail exit if any finding ≥ this
--verbose                     narrate the collapse in real time
```

## Rule catalog

Nine ways for `prabs-go` to inform you, very politely, that the code you
shipped last Friday is exactly what it looks like.

| ID | Severity | Description |
|---|---|---|
| PRABS-CPLX-001 | high    | High cyclomatic complexity — a monument to indecision |
| PRABS-FUNC-001 | medium  | Long function — a novella nobody asked for |
| PRABS-FUNC-002 | medium  | Too many parameters — struct exists, cowardice reigns |
| PRABS-NEST-001 | medium  | Deep nesting — Inception, but with fewer explosions |
| PRABS-DUP-001  | medium  | Duplicate function bodies — the sincerest form of cargo cult |
| PRABS-ERR-001  | medium  | Ignored error — brave, romantic, actionable in court |
| PRABS-ERR-002  | low     | `%v` where `%w` fits — you *almost* did the right thing |
| PRABS-NAME-001 | low     | Poor identifier naming — `foo`, `bar`, and the person who wrote them |
| PRABS-FILE-001 | low     | File exceeds size threshold — a war crime in twelve packages |

See [docs/rules.md](docs/rules.md).

## Configuration

`.prabs.yaml` — minimal indent-based YAML subset (2-space nesting). It has
fewer features than real YAML on purpose, because real YAML is itself a
security incident:

```yaml
version: 1

rules:
  PRABS-CPLX-001:
    enabled: true
    threshold: 10

exclude:
  - vendor
  - generated
  - testdata

severity:
  fail_on: high
```

## Output formats

- **text** — human-readable
- **json** — machine-readable
- **sarif** — GitHub / GitLab / Bitbucket code scanning, i.e. "human-readable
  under fluorescent light, on a Tuesday, by a very tired security engineer"

## Mutation testing

This is where `prabs-go` stops pretending. `scan` is the tool holding a
clipboard. `mutate` is the tool holding a match. `verify` is the tool
holding both, plus a stopwatch, plus your PR review.

`prabs-go verify .` clones your repo into a temp sandbox, deliberately
injects code smells, runs its own analyzer, and gives itself a score for
finding the mess it just made. It is the software equivalent of setting a
fire so you can heroically hold a bucket, then submitting the bucket-holding
for a promotion.

Sample:

```
✓ PRABS-CPLX-001
  mutation: complexity-injection
  detected: yes
...
5/5 mutations detected
Mutation Detection Score: 100%
```

100%. Every time. Not because the analyzer is that good, but because the
grader and the graded went to the same school and share a Slack DM.

Safety guarantees (tested, because "trust me" is for tools that don't ship
a mutation engine):

- The original tree is **never** mutated. We say this loudly because the
  tool spends most of its wall-clock time enthusiastically writing broken
  Go somewhere on the filesystem and it's important to know it isn't
  writing it *on top of* you.
- Sandboxes live only under `os.TempDir()` and evaporate on exit —
  assuming the process reaches its exit and not, say, the front page of
  Hacker News first.
- `Cleanup` refuses to delete anything outside the system temp dir. It has
  more restraint than most CLI tools; it has more restraint than most
  people. It has more restraint than the person who wrote it, which is,
  frankly, the whole point.
- No `git reset --hard`, `git clean -fd`, or `rm -rf` runs against your
  source. Those are levers `prabs-go` has *seen*. It has *pondered*. It
  has, for reasons purely legal and marginally moral, declined.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No findings (or all below `fail-on`) — improbable, cherish it |
| 1 | Findings ≥ `fail-on` severity — your CI is now correct and you are now sad |
| 2 | Invalid usage or config — read the README you did not read |
| 3 | Analyzer / runtime error — call an adult |

## CI/CD

Congratulations. You are about to hand `prabs-go` the keys to your pipeline.
The pipeline will not survive this. The pipeline will be *better*, but it
will not be the same pipeline. There will be a memorial.

GitHub Actions:

```yaml
- run: go install github.com/prabs/prabs-go/cmd/prabs@latest
- run: prabs-go scan . --format sarif --fail-on high > prabs.sarif
- uses: github/codeql-action/upload-sarif@v3
  with: { sarif_file: prabs.sarif }
```

GitLab CI:

```yaml
prabs:
  script:
    - prabs-go scan . --format json --fail-on high
```

Bitbucket Pipelines:

```yaml
- step:
    script:
      - prabs-go scan . --fail-on high
```

## Development

```bash
make build test vet
```

Adding a new rule: implement `analyzer.Rule` and register in
`internal/rules/register.go`. Adding a new mutator: implement
`mutation.Mutator` and append it to `mutation.All()`, then briefly consider
whether the world truly needed one more way for `prabs-go` to inflict harm.
It did. It always does.

See [docs/architecture.md](docs/architecture.md).

## FAQ

**Is `prabs-go` a weapon?**
No. `prabs-go` is a linter with, at most, delusions of grandeur and a
mutation engine.

**Then why does the README read like a nuclear-launch manual?**
Because the alternative was another cheerful "AWESOME! 🚀" README and the
world is drowning in those. You'll survive.

**Will `prabs-go` bring down my country?**
Only the technical parts. Culture, cuisine, and municipal water supply
should be unaffected.

**Where do I report a false positive?**
Into the void. Into the abyss. Alternatively, open an issue.
