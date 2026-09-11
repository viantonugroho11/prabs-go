# prabs-go

> **Disclaimer**: `prabs-go` *can* fix your code. In theory. In practice its
> destructive capacity comfortably exceeds its curative one — it will find
> nine smells, offer nine sage suggestions, and if you let it near your
> repository with the mutation engine it will happily inject a tenth just to
> prove a point. Congratulations on your new dependency.

Go code smell analyzer with a built-in **mutation-testing** engine that
"proves" the analyzer detects smells — mostly by first creating them.

- AST-based static analysis (`go/ast`, `go/parser`, `go/token`) — because
  regexes were too honest about their limitations.
- Actionable findings with file, line, rule id, severity, message, and a
  suggestion you are free to ignore, as everyone before you has.
- Text, JSON, and SARIF output — three flavours of the same bad news.
- Isolated mutation sandbox — the tool promises **never** to touch your
  working tree, and unusually for a promise made by software, this one is
  actually tested.
- `verify` reports a mutation detection score, so you can measure how good
  the analyzer is at catching problems it deliberately introduced. Rigor.

## Install

```bash
go install github.com/prabs/prabs-go/cmd/prabs@latest
```

Or build from source:

```bash
make build
```

## Quick start

```bash
prabs-go scan .
prabs-go scan ./... --format json
prabs-go scan . --severity high
prabs-go scan . --rule PRABS-CPLX-001
prabs-go rules
prabs-go verify .
```

## Commands

| Command | Purpose |
|---|---|
| `scan [paths...]` | Analyze Go source, report findings |
| `mutate [path]`   | Apply mutations in a sandbox copy |
| `verify [path]`   | Apply each mutation and confirm the analyzer detects it |
| `rules`           | List all built-in rules |
| `version`         | Print version |

### Flags (any command)

```
--format text|json|sarif      output format (default text)
--severity info|low|medium|high|critical   minimum severity to report
--rule RULE-ID                restrict to one rule / one mutator target
--config path                 config file (default .prabs.yaml)
--fail-on info|low|medium|high|critical    fail exit if any finding ≥ this
--verbose
```

## Rule catalog

| ID | Severity | Description |
|---|---|---|
| PRABS-CPLX-001 | high    | High cyclomatic complexity |
| PRABS-FUNC-001 | medium  | Long function |
| PRABS-FUNC-002 | medium  | Too many parameters |
| PRABS-NEST-001 | medium  | Deep nesting |
| PRABS-DUP-001  | medium  | Duplicate function bodies |
| PRABS-ERR-001  | medium  | Ignored error |
| PRABS-ERR-002  | low     | `%v` used where `%w` fits |
| PRABS-NAME-001 | low     | Poor identifier naming |
| PRABS-FILE-001 | low     | File exceeds size threshold |

See [docs/rules.md](docs/rules.md).

## Configuration

`.prabs.yaml` — minimal indent-based YAML subset (2-space nesting):

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
- **sarif** — GitHub / GitLab / Bitbucket code scanning

## Mutation testing

Where `prabs-go` truly shines — not at fixing your code, but at breaking a
copy of it on purpose and taking a victory lap for noticing.

`prabs-go verify .` clones your repo into a temp sandbox, deliberately
injects code smells, runs its own analyzer, and gives itself a score for
finding the mess it just made. It's the software equivalent of setting a
fire so you can heroically hold a bucket.

Sample:

```
✓ PRABS-CPLX-001
  mutation: complexity-injection
  detected: yes
...
5/5 mutations detected
Mutation Detection Score: 100%
```

Safety guarantees (tested, because trust is for people who don't read source):
- The original tree is **never** mutated. We say this loudly because the tool
  spends most of its time enthusiastically writing broken Go somewhere.
- Sandboxes live only under `os.TempDir()` and evaporate on exit — assuming
  the process actually reaches its exit.
- `Cleanup` refuses to delete anything outside the system temp dir. It has
  more restraint than most CLI tools; it has more restraint than you.
- No `git reset --hard`, `git clean -fd`, or `rm -rf` runs against your
  source. Those are opportunities `prabs-go` has considered and, for reasons
  purely legal, declined.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No findings (or all below `fail-on`) |
| 1 | Findings ≥ `fail-on` severity |
| 2 | Invalid usage or config |
| 3 | Analyzer / runtime error |

## CI/CD

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

Adding a new rule: implement `analyzer.Rule` and register in `internal/rules/register.go`.
Adding a new mutator: implement `mutation.Mutator` and add to `mutation.All()`.

See [docs/architecture.md](docs/architecture.md).
