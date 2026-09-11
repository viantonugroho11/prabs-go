# prabs-go

Go code smell analyzer with a built-in **mutation-testing** engine that proves
the analyzer actually detects the smells it claims to.

- AST/type-based static analysis (`go/ast`, `go/parser`, `go/token`)
- Actionable findings with file, line, rule id, severity, message, suggestion
- Text, JSON, and SARIF output for CI/CD
- Isolated mutation sandbox — **never** touches the user's working tree
- `verify` command reports a mutation detection score

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

`prabs-go verify .` copies the repo into a temp sandbox, injects each
mutation, runs the analyzer, and checks the intended rule fires.

Sample:

```
✓ PRABS-CPLX-001
  mutation: complexity-injection
  detected: yes
...
5/5 mutations detected
Mutation Detection Score: 100%
```

Safety guarantees (tested):
- The original tree is **never** mutated.
- Sandboxes live only under `os.TempDir()` and are removed on exit.
- `Cleanup` refuses to remove any path outside the system temp dir.
- No `git reset --hard`, `git clean -fd`, or `rm -rf` runs against the source.

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
